package main

func main() {
	server := NewApiServ(":8080")
	server.Run()
}
