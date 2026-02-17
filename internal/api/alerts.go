package api

import (
	"fmt"
	"log"
	"time"

	"eve-flipper/internal/config"
	"eve-flipper/internal/db"
	"eve-flipper/internal/engine"
)

const (
	// DefaultAlertCooldown is the minimum time between repeat alerts for the same item/metric/threshold.
	DefaultAlertCooldown = time.Hour
)

// AlertCheckResult describes whether an alert should be sent and contains necessary metadata.
type AlertCheckResult struct {
	ShouldAlert    bool
	TypeID         int32
	TypeName       string
	Metric         string
	Threshold      float64
	CurrentValue   float64
	Message        string
	CooldownActive bool
	LastAlertAt    time.Time
}

// CheckWatchlistAlerts evaluates watchlist items against scan results and determines which alerts to fire.
// Returns list of alerts that should be sent (respecting cooldown and deduplication).
func (s *Server) CheckWatchlistAlerts(userID string, results interface{}) []AlertCheckResult {
	watchlist := s.db.GetWatchlistForUser(userID)
	var alerts []AlertCheckResult

	for _, item := range watchlist {
		if !item.AlertEnabled {
			continue
		}

		metric := item.AlertMetric
		if metric == "" {
			metric = "margin_percent"
		}
		threshold := item.AlertThreshold
		if threshold <= 0 {
			threshold = item.AlertMinMargin
		}
		if threshold <= 0 {
			continue // no valid threshold
		}

		// Extract current value from results based on type
		currentValue, typeName, ok := s.extractMetricValue(item.TypeID, metric, results)
		if !ok {
			continue // item not found in results
		}

		// Check if threshold is met
		if currentValue < threshold {
			continue
		}

		// Check cooldown (deduplication)
		lastAlertTime, err := s.db.GetLastAlertTimeForUser(userID, item.TypeID, metric, threshold)
		if err != nil {
			log.Printf("[ALERT] Error checking last alert time for type %d: %v", item.TypeID, err)
			continue
		}

		cooldownActive := false
		if !lastAlertTime.IsZero() {
			elapsed := time.Since(lastAlertTime)
			if elapsed < DefaultAlertCooldown {
				cooldownActive = true
				log.Printf("[ALERT] Cooldown active for %s (last alert: %v ago)", typeName, elapsed.Round(time.Second))
				continue
			}
		}

		// Generate alert message
		message := s.formatAlertMessage(typeName, metric, threshold, currentValue)

		alerts = append(alerts, AlertCheckResult{
			ShouldAlert:    true,
			TypeID:         item.TypeID,
			TypeName:       typeName,
			Metric:         metric,
			Threshold:      threshold,
			CurrentValue:   currentValue,
			Message:        message,
			CooldownActive: cooldownActive,
			LastAlertAt:    lastAlertTime,
		})
	}

	return alerts
}

// processWatchlistAlerts evaluates alerts for a result set and sends all triggered alerts.
func (s *Server) processWatchlistAlerts(userID string, cfg *config.Config, results interface{}, scanID *int64) {
	// Desktop notifications are handled on frontend; backend processes only external channels.
	if cfg == nil || (!cfg.AlertTelegram && !cfg.AlertDiscord) {
		return
	}
	alerts := s.CheckWatchlistAlerts(userID, results)
	if len(alerts) == 0 {
		return
	}
	for _, alert := range alerts {
		if err := s.SendAlert(userID, cfg, alert, scanID); err != nil {
			log.Printf("[ALERT] Failed sending alert for type %d: %v", alert.TypeID, err)
		}
	}
}

// SendAlert sends an alert via configured channels and records it in history.
func (s *Server) SendAlert(userID string, cfg *config.Config, alert AlertCheckResult, scanID *int64) error {
	// Send via configured channels
	result := s.sendConfiguredExternalAlerts(cfg, alert.Message)

	// Record in history
	channelsSent := result.Sent
	channelsFailed := result.Failed

	entry := db.AlertHistoryEntry{
		WatchlistTypeID: alert.TypeID,
		TypeName:        alert.TypeName,
		AlertMetric:     alert.Metric,
		AlertThreshold:  alert.Threshold,
		CurrentValue:    alert.CurrentValue,
		Message:         alert.Message,
		ChannelsSent:    channelsSent,
		ChannelsFailed:  channelsFailed,
		SentAt:          time.Now().UTC().Format(time.RFC3339),
		ScanID:          scanID,
	}

	if err := s.db.SaveAlertHistoryForUser(userID, entry); err != nil {
		log.Printf("[ALERT] Failed to save alert history: %v", err)
		// Don't fail the alert send if history save fails
	}

	log.Printf("[ALERT] Sent alert for %s: %s (channels: %v)", alert.TypeName, alert.Message, channelsSent)
	return nil
}

// extractMetricValue extracts the current value for a given metric from scan results.
// Supports FlipResult, StationTrade, ContractResult, etc.
func (s *Server) extractMetricValue(typeID int32, metric string, results interface{}) (float64, string, bool) {
	// Try to cast results to different types
	switch r := results.(type) {
	case []engine.FlipResult:
		for _, item := range r {
			if item.TypeID == typeID {
				value := extractFlipMetric(item, metric)
				return value, item.TypeName, true
			}
		}
	case []engine.StationTrade:
		for _, item := range r {
			if item.TypeID == typeID {
				value := extractStationMetric(item, metric)
				return value, item.TypeName, true
			}
		}
	case []engine.ContractResult:
		// Contracts don't have type_id, skip for now
		return 0, "", false
	default:
		log.Printf("[ALERT] Unknown result type: %T", results)
		return 0, "", false
	}
	return 0, "", false
}

func extractFlipMetric(item engine.FlipResult, metric string) float64 {
	switch metric {
	case "margin_percent":
		return item.MarginPercent
	case "total_profit":
		return item.TotalProfit
	case "profit_per_unit":
		return item.ProfitPerUnit
	case "daily_volume":
		return float64(item.DailyVolume)
	default:
		return 0
	}
}

func extractStationMetric(item engine.StationTrade, metric string) float64 {
	switch metric {
	case "margin_percent":
		return item.MarginPercent
	case "total_profit":
		return item.TotalProfit
	case "profit_per_unit":
		return item.ProfitPerUnit
	case "daily_volume":
		return float64(item.DailyVolume)
	default:
		return 0
	}
}

func (s *Server) formatAlertMessage(typeName, metric string, threshold, current float64) string {
	metricLabel := metric
	switch metric {
	case "margin_percent":
		metricLabel = "Margin"
		return fmt.Sprintf("%s: %s %.2f%% >= %.2f%%", typeName, metricLabel, current, threshold)
	case "total_profit":
		metricLabel = "Total Profit"
		return fmt.Sprintf("%s: %s %.0f ISK >= %.0f ISK", typeName, metricLabel, current, threshold)
	case "profit_per_unit":
		metricLabel = "Profit/Unit"
		return fmt.Sprintf("%s: %s %.0f ISK >= %.0f ISK", typeName, metricLabel, current, threshold)
	case "daily_volume":
		metricLabel = "Daily Volume"
		return fmt.Sprintf("%s: %s %.0f >= %.0f", typeName, metricLabel, current, threshold)
	default:
		return fmt.Sprintf("%s: %s %.2f >= %.2f", typeName, metric, current, threshold)
	}
}
