package main

func main() {
	app := App{}
	app.Initialize(DbUser, DbPassword, PublicIP, Port, DbName)
	app.Run(":8080")
}