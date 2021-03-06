package nutanix

import (
	"reflect"
	"testing"
)

func TestConfig_Client(t *testing.T) {
	type fields struct {
		Endpoint string
		Username string
		Password string
		Port     string
		Insecure bool
	}
	tests := []struct {
		name    string
		fields  fields
		want    *Client
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				Endpoint: tt.fields.Endpoint,
				Username: tt.fields.Username,
				Password: tt.fields.Password,
				Port:     tt.fields.Port,
				Insecure: tt.fields.Insecure,
			}
			got, err := c.Client()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Client() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Config.Client() = %v, want %v", got, tt.want)
			}
		})
	}
}
