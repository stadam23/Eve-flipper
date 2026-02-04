import { createContext, useContext, useState, useCallback, type ReactNode } from "react";

export type Locale = "ru" | "en";

const translations = {
  ru: {
    // Header
    appTitle: "EVE Flipper",
    loginEve: "Войти через EVE",

    // Status
    sdeLoading: "SDE: загрузка...",
    sdeSystems: "систем",
    sdeTypes: "типов",
    esiApi: "ESI API",
    esiUnavailable: "ESI API: недоступен",

    // Parameters
    system: "Система",
    systemPlaceholder: "Система...",
    useCurrentLocation: "Моя локация",
    cargoCapacity: "Грузоподъёмность (m³)",
    buyRadius: "Радиус покупки (прыжки)",
    sellRadius: "Радиус продажи (прыжки)",
    minMargin: "Мин. маржа (%)",
    salesTax: "Налог продажи (%)",
    minDailyVolume: "Мин. дн. объём",
    maxInvestment: "Макс. инвестиция ISK",
    maxResults: "Лимит результатов",

    // Contract filters
    minContractPrice: "Мин. цена контракта",
    maxContractMargin: "Макс. маржа (%)",
    minPricedRatio: "Мин. доля оценки (%)",
    requireHistory: "Требовать историю",
    contractFilters: "Фильтры контрактов",
    contractFiltersHint: "Настройки защиты от скама",

    // Tabs
    tabRadius: "Флипер (радиус)",
    tabRegion: "Региональный арбитраж",
    tabContracts: "Арбитраж контрактов",

    // Buttons
    scan: "Сканировать",
    stop: "Остановить",

    // Table
    colItem: "Предмет",
    colBuyPrice: "Покупка ISK",
    colBuyStation: "Станция покупки",
    colSellPrice: "Продажа ISK",
    colSellStation: "Станция продажи",
    colMargin: "Маржа %",
    colUnitsToBuy: "Покупать",
    colAcceptQty: "Приём",
    colProfit: "Прибыль ISK",
    colDailyProfit: "Дн. прибыль",
    colProfitPerUnit: "Прибыль/шт",
    colProfitPerJump: "ISK/прыжок",
    colJumps: "Прыжки",
    colDailyVolume: "Дн. объём",
    colVelocity: "Скорость",
    colPriceTrend: "Тренд %",
    colBuyCompetitors: "Конк. покуп.",
    colSellCompetitors: "Конк. прод.",

    // Contract table
    colTitle: "Название",
    colContractPrice: "Цена контракта",
    colMarketValue: "Рыночная стоимость",
    colContractProfit: "Прибыль",
    colContractMargin: "Маржа %",
    colVolume: "Объём m³",
    colStation: "Станция",
    colItems: "Предметов",
    colContractJumps: "Прыжки",
    colContractPPJ: "ISK/прыжок",
    foundContracts: "Найдено {count} контрактов",
    scanContractsPrompt: "Нажмите «Сканировать» для поиска контрактов",

    // Route finder
    tabRoute: "Маршрут",
    routeMinHops: "Мин. хопов",
    routeMaxHops: "Макс. хопов",
    routeSettings: "Настройки маршрута",
    routeSettingsHint: "Параметры поиска маршрутов",

    // Industry
    tabIndustry: "Индустрия",
    industrySettings: "Настройки производства",
    industrySettingsHint: "Анализ цепочки производства",
    industrySelectItem: "Выберите предмет",
    industrySearchPlaceholder: "Поиск предметов для производства...",
    industryRuns: "Кол-во запусков",
    industryME: "ME (0-10)",
    industryTE: "TE (0-20)",
    industryFacilityTax: "Налог структуры (%)",
    industryStructureBonus: "Бонус структуры (%)",
    industryAnalyze: "Анализировать",
    industryMarketPrice: "Цена на рынке",
    industryBuildCost: "Стоимость производства",
    industrySavings: "Экономия",
    industryJobCost: "Стоимость джобов",
    industryTreeView: "Дерево материалов",
    industryShoppingList: "Список покупок",
    industryPrompt: "Выберите предмет и нажмите «Анализировать» для расчёта цепочки производства",
    industryNoBlueprint: "⚠ Этот предмет нельзя производить (нет blueprint). Faction/pirate корабли получают через LP store или loot.",
    routeFind: "Найти маршруты",
    routeFound: "Найдено {count} маршрутов",
    routePrompt: "Задайте параметры и нажмите «Найти маршруты»",
    routeColumn: "Маршрут",
    routeHopsCol: "Хопов",
    routeDetails: "Детали маршрута",
    routeTotalProfit: "Общая прибыль",
    routeTotalJumps: "Прыжков",
    routeJumpsUnit: "прыжков",
    routeBuy: "Купить",
    routeSell: "Продать",
    routeDeliverTo: "Везти в",

    // Table status
    foundDeals: "Найдено {count} сделок",
    scanPrompt: "Нажмите «Сканировать» для поиска сделок",
    scanStarting: "Запуск сканирования...",
    errorPrefix: "Ошибка: ",

    // Context menu
    copyItem: "Копировать предмет",
    copyBuyStation: "Копировать станцию покупки",
    copySellStation: "Копировать станцию продажи",

    // Table features
    filterPlaceholder: "Фильтр...",
    pinRow: "Закрепить",
    unpinRow: "Открепить",
    exportCSV: "Экспорт CSV",
    copyTable: "Копировать таблицу",
    clearFilters: "Сбросить фильтры",
    selected: "Выбрано: {count}",
    totalProfit: "Сумма прибыли",
    avgMargin: "Средняя маржа",
    showing: "Показано {shown} из {total}",
    pinned: "Закреплено: {count}",

    // Watchlist
    tabWatchlist: "Избранное",
    addToWatchlist: "В избранное",
    removeFromWatchlist: "Убрать из избранного",
    watchlistEmpty: "Избранное пусто",
    watchlistHint: "ПКМ на предмет → «В избранное»",
    watchlistThreshold: "Порог %",
    watchlistCurrentMargin: "Маржа %",
    watchlistCurrentProfit: "Прибыль",
    watchlistBuyAt: "Покупка",
    watchlistSellAt: "Продажа",
    watchlistAdded: "Добавлен",
    watchlistClickToEdit: "Клик для редактирования",
    watchlistTracked: "Отслеживается",
    watchlistAlerts: "Алертов",

    // Copy / Export
    copyRoute: "Копировать маршрут",
    copyTradeRoute: "Копировать маршрут (Buy → Sell)",
    copySystemAutopilot: "Копировать систему",
    copied: "Скопировано!",

    // Notifications
    alertTriggered: "Маржа {margin}% > порог {threshold}%",

    // Station Trading
    tabStation: "Стейшн Трейдинг",
    stationSettings: "Настройки станции",
    stationSettingsHint: "Параметры сканирования",
    advancedFilters: "Расширенные фильтры",
    stationSelect: "Станция",
    brokerFee: "Комиссия брокера (%)",
    colSpread: "Спред ISK",
    colROI: "ROI %",
    colBuyOrders: "Buy ордера",
    colSellOrders: "Sell ордера",
    colBuyVolume: "Buy объём",
    colSellVolume: "Sell объём",
    stationPrompt: "Выберите станцию и нажмите «Сканировать»",
    foundStationDeals: "Найдено {count} возможностей",
    noStations: "Нет станций в системе",
    loadingStations: "Загрузка станций...",
    allStations: "Все станции",
    stationRadius: "Радиус",
    colStationName: "Станция",

    // EVE Guru Station Trading metrics
    profitFilters: "Фильтры прибыли",
    riskProfile: "Профиль риска",
    bvsAndLimits: "B v S и лимиты",
    minItemProfit: "Мин. прибыль/шт",
    minDemandPerDay: "Мин. спрос/день",
    avgPricePeriod: "Период (дн.)",
    minPeriodROI: "Мин. Period ROI",
    maxPVI: "Макс. PVI",
    maxSDS: "Макс. SDS",
    bvsRatioMin: "B v S мин",
    bvsRatioMax: "B v S макс",
    limitBuyToPriceLow: "Лимит покупки по P.Low",
    flagExtremePrices: "Флаг экстрем. цен",
    colCTS: "CTS",
    colPeriodROI: "P.ROI %",
    colBuyPerDay: "Спрос/день",
    colBvS: "B v S",
    colDOS: "D.O.S.",
    colSDS: "SDS",
    colVWAP: "VWAP",
    colOBDS: "OBDS",
    colNowROI: "Now ROI",
    colCapital: "Капитал",
    highRisk: "Высокий риск",
    extremePrice: "Экстрем. цена",
    avgCTS: "Средний CTS",

    // History
    scanHistory: "История сканирований",
    historyEmpty: "Нет истории",
    noScanHistory: "История сканирований пуста",
    confirmDeleteScan: "Удалить эту запись?",
    clearOlderThanDays: "Удалить записи старше скольки дней?",
    clearedScans: "Удалено {count} записей",
    failedToLoadResults: "Не удалось загрузить результаты",
    loadToTab: "Загрузить во вкладку",
    historyBack: "Назад",
    historyResults: "Результаты",
    historyTopProfit: "Топ прибыль",
    historyTotalProfit: "Общая прибыль",
    historyDuration: "Время",
    scanParameters: "Параметры сканирования",
    resultsPreview: "Предпросмотр результатов",
    historyItems: "элементов",
    andMore: "и ещё {count}",
    historyType: "Тип",
    historyLocation: "Локация",
    historyTime: "Время",
    historyView: "Смотреть",
    historyDelete: "Удалить",
    viewResults: "Просмотр результатов",
    historyRefresh: "Обновить",
    clearOld: "Очистить",
    yesterday: "Вчера",
    radiusScan: "Радиус скан",
    regionArbitrage: "Региональный арбитраж",
    historyContracts: "Контракты",
    historyStationTrading: "Станционная торговля",
    historyRouteBuilder: "Маршруты",
    tabHistory: "История",
    loading: "Загрузка",
    colItemName: "Предмет",
    historyProfit: "Прибыль",
    historyMargin: "Маржа",

    // Character
    charViewInfo: "Информация о персонаже",
    logout: "Выйти",
    charWallet: "Кошелёк",
    charBuyOrders: "Buy ордера",
    charSellOrders: "Sell ордера",
    charTotalOrders: "Всего ордеров",
    charActiveOrders: "Активные ордера",
    charOrderType: "Тип",
    charPrice: "Цена",
    charVolume: "Объём",
    charTotal: "Всего",
    charNoOrders: "Нет активных ордеров",
    charSkillPoints: "Очки навыков",
    charOverview: "Обзор",
    charOrderHistory: "История ордеров",
    charTransactions: "Транзакции",
    charEscrow: "В залоге",
    charSellOrdersValue: "Sell ордера",
    charNetWorth: "Капитал",
    charTradingProfit: "Прибыль торговли",
    charRecentBuys: "Недавние покупки",
    charRecentSales: "Недавние продажи",
    charTxns: "транзакций",
    charAll: "Все",
    charBuy: "Покупка",
    charSell: "Продажа",
    charLocation: "Станция",
    charIssued: "Создан",
    charNoHistory: "Нет истории ордеров",
    charFulfilled: "Выполнен",
    charCancelled: "Отменён",
    charExpired: "Истёк",
    charState: "Статус",
    charFilled: "Заполнен",
    charNoTransactions: "Нет транзакций",
    charUnitPrice: "Цена/шт",
    charQty: "Кол-во",
    charDate: "Дата",
    
    // Dialogs
    dialogConfirm: "Подтвердить",
    dialogCancel: "Отмена",
    dialogDelete: "Удалить",
    dialogOk: "OK",
    
    // Metric tooltips
    metricCTSTitle: "CTS — Composite Trading Score",
    metricCTSDesc: "Комплексная оценка качества торговли. Учитывает объём, маржу, волатильность и риски.",
    metricCTSGood: "≥70",
    metricCTSBad: "<40",
    
    metricSDSTitle: "SDS — Scam Detection Score",
    metricSDSDesc: "Индикатор риска манипуляций. Высокие значения указывают на возможные скам-ордера.",
    metricSDSGood: "<30",
    metricSDSBad: "≥50",
    
    metricVWAPTitle: "VWAP — Volume Weighted Average Price",
    metricVWAPDesc: "Средневзвешенная по объёму цена за период. Показывает реальную среднюю цену сделок.",
    
    metricPVITitle: "PVI — Price Volatility Index",
    metricPVIDesc: "Индекс волатильности цены. Высокие значения означают нестабильную цену.",
    metricPVIGood: "<20%",
    metricPVIBad: ">50%",
    
    metricOBDSTitle: "OBDS — Order Book Depth Score",
    metricOBDSDesc: "Глубина книги ордеров. Показывает ликвидность рынка для данного предмета.",
    
    metricDOSTitle: "D.O.S. — Days of Supply",
    metricDOSDesc: "Дни запаса. Сколько дней текущий запас покроет спрос при средних продажах.",
    metricDOSGood: "1-7 дней",
    metricDOSBad: ">30 дней",
    
    metricBvSTitle: "B v S — Buy vs Sell Ratio",
    metricBvSDesc: "Соотношение покупок к продажам. Показывает баланс спроса и предложения.",
    metricBvSGood: "0.5-2.0",
    metricBvSBad: "<0.2 или >5",
    
    metricPeriodROITitle: "Period ROI",
    metricPeriodROIDesc: "Доходность за период. Процент прибыли относительно вложенного капитала за выбранный период.",
    
    metricNowROITitle: "Now ROI",
    metricNowROIDesc: "Текущая доходность. ROI на основе текущих цен покупки и продажи.",
  },
  en: {
    // Header
    appTitle: "EVE Flipper",
    loginEve: "Login with EVE",

    // Status
    sdeLoading: "SDE: loading...",
    sdeSystems: "systems",
    sdeTypes: "types",
    esiApi: "ESI API",
    esiUnavailable: "ESI API: unavailable",

    // Parameters
    system: "System",
    systemPlaceholder: "System...",
    useCurrentLocation: "My location",
    cargoCapacity: "Cargo Capacity (m³)",
    buyRadius: "Buy Radius (jumps)",
    sellRadius: "Sell Radius (jumps)",
    minMargin: "Min Margin (%)",
    salesTax: "Sales Tax (%)",
    minDailyVolume: "Min Daily Volume",
    maxInvestment: "Max Investment ISK",

    // Contract filters
    minContractPrice: "Min Contract Price",
    maxContractMargin: "Max Margin (%)",
    minPricedRatio: "Min Priced Ratio (%)",
    requireHistory: "Require History",
    contractFilters: "Contract Filters",
    contractFiltersHint: "Scam protection settings",
    maxResults: "Results Limit",

    // Tabs
    tabRadius: "Flipper (radius)",
    tabRegion: "Regional Arbitrage",
    tabContracts: "Contract Arbitrage",

    // Buttons
    scan: "Scan",
    stop: "Stop",

    // Table
    colItem: "Item",
    colBuyPrice: "Buy ISK",
    colBuyStation: "Buy Station",
    colSellPrice: "Sell ISK",
    colSellStation: "Sell Station",
    colMargin: "Margin %",
    colUnitsToBuy: "Buy Qty",
    colAcceptQty: "Accept Qty",
    colProfit: "Profit ISK",
    colDailyProfit: "Daily Profit",
    colProfitPerUnit: "Profit/Unit",
    colProfitPerJump: "ISK/Jump",
    colJumps: "Jumps",
    colDailyVolume: "Daily Vol",
    colVelocity: "Velocity",
    colPriceTrend: "Trend %",
    colBuyCompetitors: "Buy Comp.",
    colSellCompetitors: "Sell Comp.",

    // Contract table
    colTitle: "Title",
    colContractPrice: "Contract Price",
    colMarketValue: "Market Value",
    colContractProfit: "Profit",
    colContractMargin: "Margin %",
    colVolume: "Volume m³",
    colStation: "Station",
    colItems: "Items",
    colContractJumps: "Jumps",
    colContractPPJ: "ISK/Jump",
    foundContracts: "Found {count} contracts",
    scanContractsPrompt: "Press \"Scan\" to search for contracts",

    // Route finder
    tabRoute: "Route",
    routeMinHops: "Min hops",
    routeMaxHops: "Max hops",
    routeSettings: "Route Settings",
    routeSettingsHint: "Route search parameters",

    // Industry
    tabIndustry: "Industry",
    industrySettings: "Production Settings",
    industrySettingsHint: "Production chain analysis",
    industrySelectItem: "Select item",
    industrySearchPlaceholder: "Search items to produce...",
    industryRuns: "Runs",
    industryME: "ME (0-10)",
    industryTE: "TE (0-20)",
    industryFacilityTax: "Facility Tax (%)",
    industryStructureBonus: "Structure Bonus (%)",
    industryAnalyze: "Analyze",
    industryMarketPrice: "Market Price",
    industryBuildCost: "Build Cost",
    industrySavings: "Savings",
    industryJobCost: "Job Cost",
    industryTreeView: "Material Tree",
    industryShoppingList: "Shopping List",
    industryPrompt: "Select an item and click \"Analyze\" to calculate the production chain",
    industryNoBlueprint: "⚠ This item cannot be manufactured (no blueprint). Faction/pirate ships are obtained via LP store or loot.",
    routeFind: "Find routes",
    routeFound: "Found {count} routes",
    routePrompt: "Set parameters and press \"Find routes\"",
    routeColumn: "Route",
    routeHopsCol: "Hops",
    routeDetails: "Route details",
    routeTotalProfit: "Total profit",
    routeTotalJumps: "Jumps",
    routeJumpsUnit: "jumps",
    routeBuy: "Buy",
    routeSell: "Sell",
    routeDeliverTo: "Deliver to",

    // Table status
    foundDeals: "Found {count} deals",
    scanPrompt: "Press \"Scan\" to search for deals",
    scanStarting: "Starting scan...",
    errorPrefix: "Error: ",

    // Context menu
    copyItem: "Copy item name",
    copyBuyStation: "Copy buy station",
    copySellStation: "Copy sell station",

    // Table features
    filterPlaceholder: "Filter...",
    pinRow: "Pin",
    unpinRow: "Unpin",
    exportCSV: "Export CSV",
    copyTable: "Copy table",
    clearFilters: "Clear filters",
    selected: "Selected: {count}",
    totalProfit: "Total profit",
    avgMargin: "Avg margin",
    showing: "Showing {shown} of {total}",
    pinned: "Pinned: {count}",

    // Watchlist
    tabWatchlist: "Watchlist",
    addToWatchlist: "Add to watchlist",
    removeFromWatchlist: "Remove from watchlist",
    watchlistEmpty: "Watchlist is empty",
    watchlistHint: "Right-click item → \"Add to watchlist\"",
    watchlistThreshold: "Threshold %",
    watchlistCurrentMargin: "Margin %",
    watchlistCurrentProfit: "Profit",
    watchlistBuyAt: "Buy",
    watchlistSellAt: "Sell",
    watchlistAdded: "Added",
    watchlistClickToEdit: "Click to edit",
    watchlistTracked: "Tracked",
    watchlistAlerts: "Alerts",

    // Copy / Export
    copyRoute: "Copy route",
    copyTradeRoute: "Copy route (Buy → Sell)",
    copySystemAutopilot: "Copy system name",
    copied: "Copied!",

    // Notifications
    alertTriggered: "Margin {margin}% > threshold {threshold}%",

    // Station Trading
    tabStation: "Station Trading",
    stationSettings: "Station Settings",
    stationSettingsHint: "Scan parameters",
    advancedFilters: "Advanced Filters",
    stationSelect: "Station",
    brokerFee: "Broker fee (%)",
    colSpread: "Spread ISK",
    colROI: "ROI %",
    colBuyOrders: "Buy Orders",
    colSellOrders: "Sell Orders",
    colBuyVolume: "Buy Volume",
    colSellVolume: "Sell Volume",
    stationPrompt: "Select a station and press \"Scan\"",
    foundStationDeals: "Found {count} opportunities",
    noStations: "No stations in system",
    loadingStations: "Loading stations...",
    allStations: "All stations",
    stationRadius: "Radius",
    colStationName: "Station",

    // EVE Guru Station Trading metrics
    profitFilters: "Profit Filters",
    riskProfile: "Risk Profile",
    bvsAndLimits: "B v S & Limits",
    minItemProfit: "Min Profit/Unit",
    minDemandPerDay: "Min Demand/Day",
    avgPricePeriod: "Period (days)",
    minPeriodROI: "Min Period ROI",
    maxPVI: "Max PVI",
    maxSDS: "Max SDS",
    bvsRatioMin: "B v S Min",
    bvsRatioMax: "B v S Max",
    limitBuyToPriceLow: "Limit Buy to P.Low",
    flagExtremePrices: "Flag Extreme Prices",
    colCTS: "CTS",
    colPeriodROI: "P.ROI %",
    colBuyPerDay: "Buy/Day",
    colBvS: "B v S",
    colDOS: "D.O.S.",
    colSDS: "SDS",
    colVWAP: "VWAP",
    colOBDS: "OBDS",
    colNowROI: "Now ROI",
    colCapital: "Capital",
    highRisk: "High Risk",
    extremePrice: "Extreme Price",
    avgCTS: "Avg CTS",

    // History
    scanHistory: "Scan history",
    historyEmpty: "No history",
    noScanHistory: "No scan history yet",
    confirmDeleteScan: "Delete this scan record?",
    clearOlderThanDays: "Delete records older than how many days?",
    clearedScans: "Deleted {count} records",
    failedToLoadResults: "Failed to load results",
    loadToTab: "Load to Tab",
    historyBack: "Back",
    historyResults: "Results",
    historyTopProfit: "Top Profit",
    historyTotalProfit: "Total Profit",
    historyDuration: "Duration",
    scanParameters: "Scan Parameters",
    resultsPreview: "Results Preview",
    historyItems: "items",
    andMore: "and {count} more",
    historyType: "Type",
    historyLocation: "Location",
    historyTime: "Time",
    historyView: "View",
    historyDelete: "Delete",
    viewResults: "View Results",
    historyRefresh: "Refresh",
    clearOld: "Clear Old",
    yesterday: "Yesterday",
    radiusScan: "Radius Scan",
    regionArbitrage: "Region Arbitrage",
    historyContracts: "Contracts",
    historyStationTrading: "Station Trading",
    historyRouteBuilder: "Route Builder",
    tabHistory: "History",
    loading: "Loading",
    colItemName: "Item",
    historyProfit: "Profit",
    historyMargin: "Margin",

    // Character
    charViewInfo: "Character info",
    logout: "Logout",
    charWallet: "Wallet",
    charBuyOrders: "Buy Orders",
    charSellOrders: "Sell Orders",
    charTotalOrders: "Total Orders",
    charActiveOrders: "Active Orders",
    charOrderType: "Type",
    charPrice: "Price",
    charVolume: "Volume",
    charTotal: "Total",
    charNoOrders: "No active orders",
    charSkillPoints: "Skill Points",
    charOverview: "Overview",
    charOrderHistory: "Order History",
    charTransactions: "Transactions",
    charEscrow: "In Escrow",
    charSellOrdersValue: "Sell Orders",
    charNetWorth: "Net Worth",
    charTradingProfit: "Trading Profit",
    charRecentBuys: "Recent Buys",
    charRecentSales: "Recent Sales",
    charTxns: "txns",
    charAll: "All",
    charBuy: "Buy",
    charSell: "Sell",
    charLocation: "Station",
    charIssued: "Issued",
    charNoHistory: "No order history",
    charFulfilled: "Fulfilled",
    charCancelled: "Cancelled",
    charExpired: "Expired",
    charState: "State",
    charFilled: "Filled",
    charNoTransactions: "No transactions",
    charUnitPrice: "Unit Price",
    charQty: "Qty",
    charDate: "Date",
    
    // Dialogs
    dialogConfirm: "Confirm",
    dialogCancel: "Cancel",
    dialogDelete: "Delete",
    dialogOk: "OK",
    
    // Metric tooltips
    metricCTSTitle: "CTS — Composite Trading Score",
    metricCTSDesc: "Overall trading quality score. Considers volume, margin, volatility and risks.",
    metricCTSGood: "≥70",
    metricCTSBad: "<40",
    
    metricSDSTitle: "SDS — Scam Detection Score",
    metricSDSDesc: "Risk indicator for market manipulation. High values suggest potential scam orders.",
    metricSDSGood: "<30",
    metricSDSBad: "≥50",
    
    metricVWAPTitle: "VWAP — Volume Weighted Average Price",
    metricVWAPDesc: "Volume-weighted average price over the period. Shows the true average transaction price.",
    
    metricPVITitle: "PVI — Price Volatility Index",
    metricPVIDesc: "Price volatility index. High values indicate unstable prices.",
    metricPVIGood: "<20%",
    metricPVIBad: ">50%",
    
    metricOBDSTitle: "OBDS — Order Book Depth Score",
    metricOBDSDesc: "Order book depth. Shows market liquidity for this item.",
    
    metricDOSTitle: "D.O.S. — Days of Supply",
    metricDOSDesc: "Days of supply. How many days current inventory will cover demand at average sales.",
    metricDOSGood: "1-7 days",
    metricDOSBad: ">30 days",
    
    metricBvSTitle: "B v S — Buy vs Sell Ratio",
    metricBvSDesc: "Buy to sell ratio. Shows balance between demand and supply.",
    metricBvSGood: "0.5-2.0",
    metricBvSBad: "<0.2 or >5",
    
    metricPeriodROITitle: "Period ROI",
    metricPeriodROIDesc: "Return on investment for the period. Profit percentage relative to invested capital.",
    
    metricNowROITitle: "Now ROI",
    metricNowROIDesc: "Current return on investment. ROI based on current buy and sell prices.",
  },
} as const;

export type TranslationKey = keyof (typeof translations)["ru"];

interface I18nContextType {
  locale: Locale;
  setLocale: (locale: Locale) => void;
  t: (key: TranslationKey, params?: Record<string, string | number>) => string;
}

const I18nContext = createContext<I18nContextType>(null!);

export function I18nProvider({ children }: { children: ReactNode }) {
  const [locale, setLocale] = useState<Locale>(() => {
    const saved = localStorage.getItem("eve-flipper-locale");
    return (saved === "en" || saved === "ru") ? saved : "ru";
  });

  const changeLocale = useCallback((l: Locale) => {
    setLocale(l);
    localStorage.setItem("eve-flipper-locale", l);
  }, []);

  const t = useCallback(
    (key: TranslationKey, params?: Record<string, string | number>) => {
      let str: string = translations[locale][key] ?? key;
      if (params) {
        for (const [k, v] of Object.entries(params)) {
          str = str.replace(`{${k}}`, String(v));
        }
      }
      return str;
    },
    [locale]
  );

  return (
    <I18nContext.Provider value={{ locale, setLocale: changeLocale, t }}>
      {children}
    </I18nContext.Provider>
  );
}

export function useI18n() {
  return useContext(I18nContext);
}
