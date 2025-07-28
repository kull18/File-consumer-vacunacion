package data


type SensorDataVaccine struct {
	MeasurementUnit   string `json:"measurementUnit"`
	NameSensor        string `json:"nameSensor"`
	Information       float64    `json:"information"`
	IDVaccineBox      int    `json:"idVaccineBox"`
	IDSensorsVaccine  int    `json:"idSensorsVaccine"`
}