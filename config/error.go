package config

import "log"

// FailOnLoad logs when error occurred on loading config
func FailOnLoad(err error, name string) {
	if err == nil {
		log.Println("Failed to load config: " + name)
		return
	}
	log.Println("Failed to load config: " + name + " : " + err.Error())
}

// EmptyOnLoad logs when parameter got empty
func EmptyOnLoad(name string, useDefault bool, defaultVal string) {
	if useDefault {
		if defaultVal == "" {
			log.Println("Empty parameter: " + name + " Use default value: <empty>")
			return
		}
		log.Println("Empty parameter: " + name + " Use default value: " + defaultVal)
		return
	}
	log.Println("Empty parameter: " + name)
}

// FatalOnLoad logs when fatal error occurred on loading config
func FatalOnLoad(err error, name string) {
	if err == nil {
		log.Fatalln("Failed to load config: " + name)
		return
	}
	log.Fatalln("Failed to load config: " + name + " : " + err.Error())
}
