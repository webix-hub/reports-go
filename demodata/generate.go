package demodata

import (
	"math"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"time"
)

var cafeNames = []string{
	"City Stacks Coffee",
	"The Beanery",
	"Caffeine Machine",
	"Cafe Connections",
	"Espresso Love",
	"Jumpin' Beans Cafe",
	"Aroma Mocha",
	"The Friendly Bean",
	// "The Roasted Bean",
	// "Beautiful Beans",
	// "Capital Café",
	// "Cheers Cafe",
	// "Brew Together",
	// "The Teabar",
	// "Bistro at the Point",
	// "Diva Espresso",
	// "Flavored Cafeteria",
	// "Grand Cafeteria",
	// "Heavenly Coffee",
	// "Love You a Latte",
}

var products = [][]string{
	{ "Espresso", "coffee" },
	{ "Cappuccino", "coffee" },
	{ "Mocha", "coffee" },
	{ "Latte", "coffee" },
	{ "Green Tea", "tea" },
	{ "Black Tea", "tea" },
	{ "Cheescake", "dessert" },
	{ "Cherry Pie", "dessert" },
	{ "Cookie", "dessert" },
	{ "Pancakes", "dessert" },
}

var regions = []string{"Brooklyn", "Queens", "Manhattan", "Bronx"}

var firstName = []string{
	"Alda",
	"Alejandra",
	"Alva",
	"Alverta",
	"Amira",
	"Ana",
	"Briana",
	"Bryana",
	"Carrie",
	"Catharine",
	"Daniella",
	"Derek",
	"Deshawn",
	"Elvie",
	"Ena",
	"Erick",
	"Florine",
	"Gussie",
	"Ignacio",
	"Imani",
	"Jalyn",
	"Jermain",
	"Jeromy",
	"Judge",
	"Justina",
	"Kaylee",
	"Leola",
	"Logan",
	"Lois",
	"Louvenia",
	"Ludwig",
	"Luella",
	"Mara",
	"Nedra",
	"Noelia",
	"Precious",
	"Remington",
	"Ryder",
	"Sabrina",
	"Sandra",
	"Shawn",
	"Sherwood",
	"Sigrid",
	"Telly",
	"Theron",
	"Trenton",
}

var lastName = []string {
	"Bayer",
	"Beahan",
	"Bradtke",
	"Brakus",
	"Cassin",
	"Dickens",
	"Douglas",
	"DuBuque",
	"Ebert",
	"Feeney",
	"Feeney",
	"Green",
	"Hahn",
	"Hane",
	"Hansen",
	"Hayes",
	"Hegmann",
	"Hermiston",
	"Hilll",
	"Hills",
	"Howe",
	"Jakubowski",
	"Kautzer",
	"Kuphal",
	"Maggio",
	"McGlynn",
	"Medhurst",
	"Osinski",
	"Pfannerstill",
	"Raylord",
	"Reichert",
	"Rice",
	"Romaguera",
	"Ruecker",
	"Ryan",
	"Schultz",
	"Shields",
	"Sipes",
	"Sporer",
	"Steuber",
	"Stracke",
	"Tremblay",
	"Wilkinson",
	"Wisozk",
	"Wisozk",
	"Wunsch",
}

var mailDomain = []string {
	"gmail.com",
	"outlook.com",
	"yahoo.com",
	"tut.by",
}

var jobPrefix = []string {
	"","","","","","","","","","",
	"Chief ",
	"Regional ",
}

var jobName = []string {
	"Developer",
	"Developer",
	"Developer",
	"Developer",
	"Manager",
	"Manager",
	"Producer",
	"Designer",
	"Technician",
	"Technician",
	"Assistant",
	"Engineer",
	"Engineer",
	"Engineer",
}

var address = []string {
	"Crest",
	"Loaf",
	"Stream",
	"Corners",
	"Expressway",
	"Isle",
	"Bypass",
	"Rapids",
	"Route",
	"Union",
	"Squares",
	"Summit",
	"Camp",
	"Park",
	"Streets",
	"Fork",
	"Course",
	"Flats",
	"Circles",
	"Keys",
	"Locks",
	"Junctions",
	"Unions",
	"Well",
	"Spur",
	"Knolls",
	"Turnpike",
	"Vista",
	"Loaf",
	"Squares",
	"Ridge",
}

func randomName() string {
	return firstName[random(0, len(firstName))] + " " + lastName[random(0, len(lastName))]
}

func randomEmail(name string) string {
	return strings.Replace(name, " ", ".", -1)+"@"+mailDomain[randomNormal(0, len(mailDomain), 0)]
}

func randomAddress() string {
	return address[random(0, len(address))] + " " + strconv.Itoa(random(6849, 59148))
}

func randomJob() string {
	return jobPrefix[random(0, len(jobPrefix))] + jobName[random(0, len(jobName))];
}

func random(f, t int) int {
	ff := float64(f)
	tt := float64(t)
	return int(math.Floor(rand.Float64()*(tt-ff)+ff));
}

func randomPrice(f, t int) float64 {
	return float64(random(f*100,t*100))/100;
}

func randomNormal(f, t, n int) int {
	rng := rand.NormFloat64()

	left := float64(n-f)
	right := float64(t-n)
	dev := math.Max(left, right) / 3

	v := int(math.Round(rng*dev + float64(n)));
	if v >= f && v < t {
		return v
	}

	return randomNormal(f, t, n)
}

func randomDate(from time.Time, to int) time.Time {
	fr := int(from.Unix())
	t := time.Unix(int64(random(fr, fr+to)), 0)
	d := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0,0, t.Location())
	return d
}


const DAY = 60*60*24
const MONTH = DAY*31
const YEAR = MONTH*12

type Place struct {
	ID int
	Name string
	Region string
	Created time.Time

	Prefer int
	Customers float64
	Grow float64
}

type Customer struct {
	ID int
	Name string
	Email string
	Address string
	Job string

	Age int
	Prefer int
}

type Sale struct {
	ID int
	When time.Time
	PlaceID int
	Count int
	ProductID int
	Total float64
	CustomerID int
	Type int
}

type Product struct {
	ID int
	Name string
	Type string
	Price float64
}

func generate(names [][]string, cafeNames []string, regions []string, cCount int) ([]Place, []Customer, []Sale, []Product) {
	cOut := make([]Place, 0, 0)
	pOut := make([]Product, 0, 0)
	sOut := make([]Sale, 0, 0)
	csOut := make([]Customer, 0, 0)

	for i, n := range names {
		pOut = append(pOut, Product{
			ID: i+1,
			Name: n[0],
			Type: n[1],
			Price: randomPrice(3,12),
		})
	}

	for i, n := range cafeNames {
		start, _ := time.Parse("2006-01-02", "2018-01-01")
		cOut = append(cOut, Place{
			ID: i+1,
			Name: n,
			Region: regions[random(0, 4)],
			Created: randomDate(start, 2*YEAR),

			Prefer: random(0, len(pOut)),
			Customers: 1,
			Grow: randomPrice(3, 15),
		});

		sort.Slice(cOut, func(i, j int) bool {
			return cOut[i].Created.After(cOut[j].Created)
		})
		for i := range cOut {
			cOut[i].ID = i+1
		}
	}

	for i:=1; i<=cCount; i+=1 {
		name := randomName()
		csOut = append(csOut, Customer{
			ID:i,
			Name: name,
			Email:randomEmail(name),
			Age: randomNormal(19, 68, 26),
			Job: randomJob(),
			Address: randomAddress(),

			Prefer: random(0, len(pOut)),
		});
	}


	start := cOut[0].Created
	end, _ := time.Parse("2006-01-02", "2020-10-30")
	for start.Before(end) {
		for i, cafe := range cOut {
			if !cafe.Created.After(start) && rand.Float64() > 0.5 {
				// add sales
				// increase/decrease customers
				orders := random(int(math.Floor(cafe.Customers/5)), int(math.Floor(cafe.Customers)))
				for j:=0; j<orders; j+=1 {
					count := random(1, 3);
					var p *Product
					if rand.Float64() > 0.75 {
						p = &pOut[cafe.Prefer]
					} else {
						p = &pOut[random(0, len(pOut))]
					}

					var customers []Customer
					if rand.Float64() > 0.75 {
						customers = make([]Customer, 0, len(csOut))
						for _, cs := range csOut {
							if cs.Prefer == p.ID {
								customers = append(customers, cs)
							}
						}
					} else {
						customers = csOut
					}

					if len(customers) == 0 {
						continue
					}

					c := customers[random(0, len(customers))]
					sType := 1
					if rand.Float64() > 0.66 {
						sType = 2
					}

					sOut = append(sOut, Sale{
						ID: len(sOut)+1,
						When: start,
						PlaceID: cafe.ID,
						Count:count,
						ProductID: p.ID,
						Total: p.Price*float64(count),
						CustomerID: c.ID,
						Type: sType,
					})
				}

				nextCustomers := cafe.Customers + float64(random(0, 100))/100*cafe.Grow
				cOut[i].Customers = nextCustomers
				if nextCustomers >= 12 {
					cOut[i].Customers = 10;
					cOut[i].Grow = 0;
					if rand.Float64() > 0.95 {
						cOut[i].Grow = -1;
					}
				}
			}
		}
		start = start.AddDate(0, 0, 1);
	}

	return cOut, csOut, sOut, pOut
}

func defGen() ([]Place, []Customer, []Sale, []Product) {
	return generate(products, cafeNames, regions, 300)
}