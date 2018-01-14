package main

// import "os"

func main() {
	a := App{}
	/*
	a.Initialize(
		os.Getenv("APP_DB_USERNAME"),
		os.Getenv("APP_DB_PASSWORD"),
		os.Getenv("APP_DB_NAME"))
	*/
	a.Initialize("postgres", "postgres", "postgres", "disable")

	a.Run(":8080")
}