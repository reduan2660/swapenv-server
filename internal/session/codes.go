package session

import "math/rand"

var codes = []string{
	"Padma",
	"Jamuna",
	"Meghna",
	"Surma",
	"Teesta",
	"Feni",
	"Gorai",
	"Rupsha",
	"Halda",
	"Atrai",
	"Buriganga",
	"Karnaphuli",
	"Madhumati",
	"Sitalakhya",
	"Dhaleshwari",
	"Ichamati",
}

func generateCodeName() string {
	return codes[rand.Intn(len(codes))]
}
