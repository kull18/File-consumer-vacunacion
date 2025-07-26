package data


type SensorDataVaccine struct {
	MeasurementUnit   string `json:"measurementUnit"`
	NameSensor        string `json:"nameSensor"`
	Information       int    `json:"information"`
	IDVaccineBox      int    `json:"idVaccineBox"`
	IDSensorsVaccine  int    `json:"idSensorsVaccine"`
}