package dao

import (
	"goim-demo/pkg/gerrors"
)

func (d *Dao) Publish(topic string, bytes []byte) error {
	_, err := d.RdsCli.Publish(topic, bytes).Result()
	if err != nil {
		return gerrors.WrapError(err)
	}
	return nil
}
