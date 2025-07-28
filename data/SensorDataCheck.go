package data


type SensorDataCheck struct {
	MeasurementUnit    string `json:"measurementUnit"`
	NameSensor         string `json:"nameSensor"`
	Information        float64 `json:"information"`
	UserCivilIDUserCivil int    `json:"UserCivil_idUserCivil"`
}
