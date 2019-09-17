package configs

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

//Server - Dados do servidor
type Server struct {
	Port string
}

//Configuration - Configuração geral da aplicação
type Configuration struct {
	Server Server
	App    App
}

//Dbms struct
type Dbms struct {
	URI string
}

//App struct
type App struct {
	Passkey string
	Dbms    Dbms
	Elastic Elastic
}

//Elastic Config Structure
type Elastic struct {
	URL      string
	Username string
	Password string
}

//ReadConfig inicia as configurações
func ReadConfig(configFile string) (*Configuration, error) {

	var configuration Configuration

	viper.SetConfigName(configFile)
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/")
	viper.AddConfigPath("/tmp/")
	viper.AddConfigPath("$HOME/")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("Error reading config file, %s", err)
	}
	err := viper.Unmarshal(&configuration)
	if err != nil {
		return nil, fmt.Errorf("Error unmarshal config file, %s", err)
	}
	return &configuration, nil

}
