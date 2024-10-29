package constants

type CTX_CONST string

const (
	CONTEXT_KEY_LOGGER CTX_CONST = "Context_Key_Logger"
	CONTEXT_KEY_CACHE  CTX_CONST = "Context_Key_Cache"
)

// Define time zones as constants
const (
	TimezonePacificMidway      = "Pacific/Midway"
	TimezoneAmericaAdak        = "America/Adak"
	TimezoneAmericaAnchorage   = "America/Anchorage"
	TimezoneAmericaLosAngeles  = "America/Los_Angeles"
	TimezoneAmericaDenver      = "America/Denver"
	TimezoneAmericaChicago     = "America/Chicago"
	TimezoneAmericaNewYork     = "America/New_York"
	TimezoneAmericaSaoPaulo    = "America/Sao_Paulo"
	TimezoneAmericaBuenosAires = "America/Argentina/Buenos_Aires"
	TimezoneEuropeLondon       = "Europe/London"
	TimezoneEuropeParis        = "Europe/Paris"
	TimezoneEuropeBerlin       = "Europe/Berlin"
	TimezoneEuropeMoscow       = "Europe/Moscow"
	TimezoneAsiaDubai          = "Asia/Dubai"
	TimezoneAsiaKolkata        = "Asia/Kolkata"
	TimezoneAsiaSingapore      = "Asia/Singapore"
	TimezoneAsiaTokyo          = "Asia/Tokyo"
	TimezoneAsiaSeoul          = "Asia/Seoul"
	TimezoneAsiaShanghai       = "Asia/Shanghai"
	TimezoneAustraliaSydney    = "Australia/Sydney"
	TimezonePacificAuckland    = "Pacific/Auckland"
	TimezoneAfricaJohannesburg = "Africa/Johannesburg"
	TimezoneAfricaCairo        = "Africa/Cairo"
	TimezoneAfricaNairobi      = "Africa/Nairobi"
	TimezoneAmericaToronto     = "America/Toronto"
	TimezoneAmericaVancouver   = "America/Vancouver"
	TimezoneEuropeRome         = "Europe/Rome"
	TimezoneAsiaRiyadh         = "Asia/Riyadh"
)

// TimezonesSlice contains all defined time zones for easy access
var TimezonesSlice = []string{
	TimezonePacificMidway,
	TimezoneAmericaAdak,
	TimezoneAmericaAnchorage,
	TimezoneAmericaLosAngeles,
	TimezoneAmericaDenver,
	TimezoneAmericaChicago,
	TimezoneAmericaNewYork,
	TimezoneAmericaSaoPaulo,
	TimezoneAmericaBuenosAires,
	TimezoneEuropeLondon,
	TimezoneEuropeParis,
	TimezoneEuropeBerlin,
	TimezoneEuropeMoscow,
	TimezoneAsiaDubai,
	TimezoneAsiaKolkata,
	TimezoneAsiaSingapore,
	TimezoneAsiaTokyo,
	TimezoneAsiaSeoul,
	TimezoneAsiaShanghai,
	TimezoneAustraliaSydney,
	TimezonePacificAuckland,
	TimezoneAfricaJohannesburg,
	TimezoneAfricaCairo,
	TimezoneAfricaNairobi,
	TimezoneAmericaToronto,
	TimezoneAmericaVancouver,
	TimezoneEuropeRome,
	TimezoneAsiaRiyadh,
}
