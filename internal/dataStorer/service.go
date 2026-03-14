package datastorer

type StoreDataService struct{}

func (s *StoreDataService) StoreData(data []*SensorData) error {
	// Implementation for storing sensor data
	return nil
}

func NewStoreDataService() *StoreDataService {
	return &StoreDataService{}
}
