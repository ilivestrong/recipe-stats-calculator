package config

type (
	DeliveryTimeWindow struct {
		StartTime string
		EndTIme   string
	}
	Options struct {
		FixtureFilePath     string
		SearchByRecipeNames []string
		SearchByPostCode    string
		DeliveryTimeWindow  DeliveryTimeWindow
	}
	// ApplyOptions func(*Options)
)
