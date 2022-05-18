package dao

func (d *Dao) Publish(topic string, bytes []byte) error {
	_, err := d.RdsCli.Publish(topic, bytes).Result()
	if err != nil {
		return err
	}
	return nil
}
