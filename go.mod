module github.com/SENERGY-Platform/analytics-fog-master

go 1.21.3

//replace github.com/SENERGY-Platform/analytics-fog-lib => ../analytics-fog-lib

require (
	github.com/SENERGY-Platform/analytics-fog-lib v1.0.15
	github.com/SENERGY-Platform/go-service-base v0.13.0
	github.com/eclipse/paho.mqtt.golang v1.4.3
	github.com/joho/godotenv v1.5.1
	github.com/nanobox-io/golang-scribble v0.0.0-20190309225732-aa3e7c118975
	github.com/y-du/go-log-level v0.2.3
)

require (
	github.com/SENERGY-Platform/go-service-base/util v0.14.0 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/jcelliott/lumber v0.0.0-20160324203708-dd349441af25 // indirect
	github.com/y-du/go-env-loader v0.5.1 // indirect
	golang.org/x/net v0.15.0 // indirect
	golang.org/x/sync v0.3.0 // indirect
)
