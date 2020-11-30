package demodata

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"strings"
)

func InitDB(db *sqlx.DB) error {
	fmt.Println("Generating demo data")
	places, customers, sales, products := defGen()
	fmt.Printf("\t places: %d\n", len(places))
	fmt.Printf("\t products: %d\n", len(products))
	fmt.Printf("\t customers: %d\n", len(customers))
	fmt.Printf("\t sales: %d\n", len(sales))

	fmt.Print("Updating database... ")
	// drop old data, if any
	db.Exec(`DROP TABLE places`)
	db.Exec(`DROP TABLE persons`)
	db.Exec(`DROP TABLE products`)
	db.Exec(`DROP TABLE sales`)



	_, err := db.Exec(`CREATE TABLE persons (
  		id INT AUTO_INCREMENT PRIMARY KEY,
  		name VARCHAR(255) NOT NULL,
  		email VARCHAR(255) NOT NULL,
  		age INT DEFAULT 0,
  		job VARCHAR(255) NOT NULL,
  		address VARCHAR(255) NOT NULL
  	)`);
	if err != nil {

		return err
	}

	_, err = db.Exec(`CREATE TABLE places (
  		id INT AUTO_INCREMENT PRIMARY KEY,
  		name VARCHAR(255) NOT NULL,
  		region VARCHAR(255) NOT NULL,
  		created DATETIME NOT NULL
  	)`);
	if err != nil {
		return err
	}

	_, err = db.Exec(`CREATE TABLE products (
  		id INT AUTO_INCREMENT PRIMARY KEY,
  		name VARCHAR(255) NOT NULL,
  		type VARCHAR(255) NOT NULL,
  		price DECIMAL(8,2) DEFAULT 0
  	)`);
	if err != nil {
		return err
	}

	_, err = db.Exec(`CREATE TABLE sales (
  		id INT AUTO_INCREMENT PRIMARY KEY,
  		saledate DATETIME NOT NULL,
  		place_id INT DEFAULT 0,
  		count INT DEFAULT 0,
  		product_id INT DEFAULT 0,
  		total DECIMAL(8,2) DEFAULT 0,
  		customer_id INT DEFAULT 0,
  		type INT DEFAULT 0
  	)`);
	if err != nil {
		return err
	}

	for _, r := range places {
		tStr := r.Created.String()
		_, err := db.Exec("INSERT INTO places(name, region, created) VALUES(?, ?, ?)", r.Name, r.Region, tStr[0:10])
		if err != nil {
			return err
		}
	}

	for _, r := range customers {
		_, err := db.Exec("INSERT INTO persons(name, email, age, job, address) VALUES(?, ?, ?, ?, ?)", r.Name, r.Email, r.Age, r.Job, r.Address)
		if err != nil {
			return err
		}
	}

	for _, r := range products {
		_, err := db.Exec("INSERT INTO products(name, type, price) VALUES(?, ?, ?)", r.Name, r.Type, r.Price)
		if err != nil {
			return err
		}
	}

	count := len(sales)
	for i :=0; i<count; i+=100 {
		fStr := make([]string, 1, 101)
		params := make([]interface{}, 0, 100*7)
		fStr[0] = "INSERT INTO sales(saledate, place_id, count, product_id, total, customer_id, type) VALUES(?, ?, ?, ?, ?, ?, ?)";

		for j:=0; j<100; j+=1 {
			ind := i+j
			if ind >= count {
				continue
			}

			r := sales[ind]
			tStr := r.When.String()
			fStr = append(fStr, "(?, ?, ?, ?, ?, ?, ?)")
			params = append(params, tStr[0:10], r.PlaceID, r.Count, r.ProductID, r.Total, r.CustomerID, r.Type)
		}

		_, err := db.Exec(strings.Join(fStr[0:len(fStr)-1], ","), params...)
		if err != nil {
			return err
		}
	}

	fmt.Print("done\n\n")
	return nil
}

func InitReports(db *sqlx.DB) error {
	queries := []string{
		`INSERT INTO reports.queries (id, text, name) VALUES (23, '{"glue":"and","rules":[{"field":"sales.saledate","type":"date","condition":{"filter":"2019-12-31T21:00:00.000Z","type":"greaterOrEqual"},"includes":[]},{"field":"sales.saledate","type":"date","condition":{"filter":"2020-12-30T21:00:00.000Z","type":"lessOrEqual"},"includes":[]}]}', 'Sales 2020');`,
		`INSERT INTO reports.queries (id, text, name) VALUES (24, '{"glue":"and","rules":[{"field":"products.type","type":"text","condition":{"filter":"","type":"contains"},"includes":["dessert"]}]}', 'Dessert');`,

		`INSERT INTO reports.modules (id, name, text, updated) VALUES (31, 'All customers', '{"desc":"Created on 12/02/2020","data":"persons","joins":[],"query":"","columns":[{"id":"persons.name","name":"Name","type":"text","ref":"","mid":"persons","model":"Persons","meta":{}},{"id":"persons.email","name":"Email","type":"text","ref":"","mid":"persons","model":"Persons","meta":{}},{"id":"persons.age","name":"Age","type":"number","ref":"","mid":"persons","model":"Persons","meta":{},"width":77},{"id":"persons.job","name":"Job Title","type":"text","ref":"","mid":"persons","model":"Persons","meta":{},"width":156},{"id":"persons.address","name":"Address","type":"","ref":"","mid":"persons","model":"Persons","meta":{}}],"group":null,"meta":{"freeze":0},"sort":[{"id":"persons.name","mod":"asc"}],"type":"table"}', '2020-12-02 13:57:29');`,
		`INSERT INTO reports.modules (id, name, text, updated) VALUES (32, 'Sales by year', '{"desc":"Created on 12/02/2020","data":"persons","joins":[{"sid":"persons","tid":"sales","tf":"customer_id","id":"sales/customer_id//persons"}],"query":"","columns":[{"id":"year.sales.saledate","name":"Sale Date","type":"text","ref":"","mid":"sales","model":"Sales","meta":{}},{"id":"sum.sales.count","name":"Sum Count","type":"number","meta":{"name":"Items sold","header":"none"},"model":""},{"id":"sum.sales.total","name":"Sum Total","type":"number","meta":{"name":"Total Sales","header":"none"},"model":""},{"id":"avg.persons.age","name":"Average Age","type":"number","meta":{"name":"average Age of Customer","header":"none"},"model":""}],"group":{"by":[{"id":"sales.saledate","mod":"year"}],"columns":[{"op":"sum","id":"sales.count","name":"Sum Count","type":"number"},{"op":"avg","id":"persons.age","name":"Average Age","type":"number"},{"op":"sum","id":"sales.total","name":"Sum Total","type":"number"}]},"meta":{"freeze":0},"sort":null,"type":"table"}', '2020-12-02 14:13:42');`,
		`INSERT INTO reports.modules (id, name, text, updated) VALUES (33, 'Sales by month', '{"desc":"Created on 12/02/2020","data":"persons","joins":[{"sid":"persons","tid":"sales","tf":"customer_id","id":"sales/customer_id//persons"}],"query":"","columns":[{"id":"yearmonth.sales.saledate","name":"Sale Date","type":"text","ref":"","mid":"sales","model":"Sales","meta":{}},{"id":"sum.sales.count","name":"Sum Count","type":"number","meta":{},"model":""},{"id":"sum.sales.total","name":"Sum Total","type":"number","meta":{},"model":""}],"group":{"by":[{"id":"sales.saledate","mod":"yearmonth"}],"columns":[{"op":"sum","id":"sales.count","name":"Sum Count","type":"number"},{"op":"sum","id":"sales.total","name":"Sum Total","type":"number"}]},"meta":{"chart":{"chartType":"splineArea","labelColumn":"yearmonth.sales.saledate","updatedSeries":null,"seriesFrom":"columns","baseColumn":null,"dataColumn":null,"series":[{"id":"sum.sales.count","name":"Sum Count","model":null,"show":true,"meta":{"name":"Sum Count","color":"#FF9700"},"$sub":"<div></div>"},{"id":"sum.sales.total","name":"Sum Total","model":null,"show":true,"meta":{"name":"Sum Total","color":"#4CB050"},"$sub":"<div></div>"}],"legend":{"layout":"x"},"axises":{"x":{"title":"","color":"#edeff0","lineColor":"#edeff0","lines":true,"verticalLabels":true},"y":{"title":"","color":"#edeff0","lineColor":"#edeff0","lines":true,"logarithmic":false}},"legendPosition":"bottom"}},"sort":null,"type":"chart"}', '2020-12-02 14:13:32');`,
		`INSERT INTO reports.modules (id, name, text, updated) VALUES (34, 'Most popular desserts', '{"desc":"Created on 12/02/2020","data":"products","joins":[{"sid":"products","tid":"sales","tf":"product_id","id":"sales/product_id//products"}],"query":"{\\"glue\\":\\"and\\",\\"rules\\":[{\\"field\\":\\"products.type\\",\\"type\\":\\"text\\",\\"condition\\":{\\"filter\\":\\"\\",\\"type\\":\\"contains\\"},\\"includes\\":[\\"dessert\\"]}]}","columns":[{"id":"products.name","name":"Name","type":"text","ref":"","mid":"products","model":"Products","meta":{},"width":117},{"id":"sum.sales.count","name":"Sum Count","type":"number","meta":{},"model":""}],"group":{"by":[{"id":"products.id"},{"id":"products.name"}],"columns":[{"op":"sum","id":"sales.count","name":"Sum Count","type":"number"}]},"meta":{"freeze":0},"sort":[{"id":"sum.sales.count","mod":"desc"}],"type":"table"}', '2020-12-02 14:19:56');`,
		`INSERT INTO reports.modules (id, name, text, updated) VALUES (35, 'Sales by product 2020', '{"desc":"Created on 12/02/2020","data":"products","joins":[{"sid":"products","tid":"sales","tf":"product_id","id":"sales/product_id//products"}],"query":"{\\"glue\\":\\"and\\",\\"rules\\":[{\\"field\\":\\"sales.saledate\\",\\"type\\":\\"date\\",\\"condition\\":{\\"filter\\":\\"2019-12-31T21:00:00.000Z\\",\\"type\\":\\"greaterOrEqual\\"},\\"includes\\":[]},{\\"field\\":\\"sales.saledate\\",\\"type\\":\\"date\\",\\"condition\\":{\\"filter\\":\\"2020-12-30T21:00:00.000Z\\",\\"type\\":\\"lessOrEqual\\"},\\"includes\\":[]}]}","columns":[{"id":"sum.sales.total","name":"Sum Total","type":"number","meta":{},"model":""},{"id":"products.name","name":"Name","type":"text","ref":"","mid":"products","model":"Products","meta":{}}],"group":{"by":[{"id":"products.id"},{"id":"products.name"}],"columns":[{"op":"sum","id":"sales.total","name":"Sum Total","type":"number"}]},"meta":{"value":"sum.sales.total","label":"products.name","color":"sum.sales.total"},"sort":null,"type":"heatmap"}', '2020-12-02 14:19:06');`,
		`INSERT INTO reports.modules (id, name, text, updated) VALUES (36, 'Sales of tea 2019', '{"desc":"Created on 12/02/2020","data":"sales","joins":[{"sid":"sales","tid":"products","sf":"product_id","id":"sales/product_id//products"}],"query":"{\\"glue\\":\\"and\\",\\"rules\\":[{\\"field\\":\\"products.name\\",\\"type\\":\\"text\\",\\"condition\\":{\\"filter\\":\\"tea\\",\\"type\\":\\"contains\\"},\\"includes\\":[]}]}","columns":[{"id":"sales.id","name":"ID","type":"number","ref":"","mid":"sales","model":"Sales","meta":{},"key":true,"width":69},{"id":"sales.saledate","name":"Sale Date","type":"date","ref":"","mid":"sales","model":"Sales","meta":{},"width":121},{"id":"sales.place_id","name":"Place","type":"reference","ref":"places","mid":"sales","model":"Sales","meta":{},"width":174},{"id":"sales.count","name":"Count","type":"number","ref":"","mid":"sales","model":"Sales","meta":{},"width":74},{"id":"sales.product_id","name":"Product","type":"reference","ref":"products","mid":"sales","model":"Sales","meta":{},"width":109},{"id":"sales.total","name":"Total","type":"number","ref":"","mid":"sales","model":"Sales","meta":{},"width":88},{"id":"sales.customer_id","name":"Customer","type":"reference","ref":"persons","mid":"sales","model":"Sales","meta":{}},{"id":"sales.type","name":"Payment","type":"picklist","ref":"regions","mid":"sales","model":"Sales","meta":{}}],"group":null,"meta":{"freeze":0},"sort":[{"id":"sales.place_id","mod":"asc"},{"id":"sales.saledate","mod":"asc"}],"type":"table"}', '2020-12-02 14:22:42');`,
		`INSERT INTO reports.modules (id, name, text, updated) VALUES (37, 'Latte sales in different units', '{"desc":"Created on 12/02/2020","data":"places","joins":[{"sid":"places","tid":"sales","tf":"place_id","id":"sales/place_id//places"},{"sid":"sales","tid":"products","sf":"product_id","id":"sales/product_id//products"}],"query":"{\\"glue\\":\\"and\\",\\"rules\\":[{\\"field\\":\\"products.name\\",\\"type\\":\\"text\\",\\"condition\\":{\\"filter\\":\\"\\",\\"type\\":\\"contains\\"},\\"includes\\":[\\"Latte\\"]}]}","columns":[{"id":"yearmonth.sales.saledate","name":"Sale Date","type":"text","ref":"","mid":"sales","model":"Sales","meta":{}},{"id":"places.name","name":"Name","type":"text","ref":"","mid":"places","model":"Places","meta":{}},{"id":"sum.sales.total","name":"Sum Total","type":"number","meta":{},"model":""}],"group":{"by":[{"id":"places.id"},{"id":"places.name"},{"id":"sales.saledate","mod":"yearmonth"}],"columns":[{"op":"sum","id":"sales.total","name":"Sum Total","type":"number"}]},"meta":{"chart":{"chartType":"stackedArea","labelColumn":"yearmonth.sales.saledate","updatedSeries":null,"seriesFrom":"rows","baseColumn":"places.name","dataColumn":"sum.sales.total","series":[{"id":"Caffeine Machine","name":"Caffeine Machine","model":null,"show":true,"meta":{"name":"Caffeine Machine","color":"#E8B161"},"$sub":"<div></div>"},{"id":"The Beanery","name":"The Beanery","model":null,"show":true,"meta":{"name":"The Beanery","color":"#4CB050"},"$sub":"<div></div>"},{"id":"Aroma Mocha","name":"Aroma Mocha","model":null,"show":true,"meta":{"name":"Aroma Mocha","color":"#00BCD4"},"$sub":"<div></div>"},{"id":"Jumpin'' Beans Cafe","name":"Jumpin'' Beans Cafe","model":null,"show":true,"meta":{"name":"Jumpin'' Beans Cafe","color":"#8994D6"},"$sub":"<div></div>"},{"id":"Cafe Connections","name":"Cafe Connections","model":null,"show":true,"meta":{"name":"Cafe Connections","color":"#C464D6"},"$sub":"<div></div>"},{"id":"City Stacks Coffee","name":"City Stacks Coffee","model":null,"show":true,"meta":{"name":"City Stacks Coffee","color":"#DD7871"},"$sub":"<div></div>"},{"id":"The Friendly Bean","name":"The Friendly Bean","model":null,"show":true,"meta":{"name":"The Friendly Bean","color":"#F2E474"},"$sub":"<div></div>"},{"id":"Espresso Love","name":"Espresso Love","model":null,"show":true,"meta":{"name":"Espresso Love","color":"#5CD1C5"},"$sub":"<div></div>"}],"legend":{"layout":"x"},"axises":{"x":{"title":"","color":"#edeff0","lineColor":"#edeff0","lines":true,"verticalLabels":true},"y":{"title":"","color":"#edeff0","lineColor":"#edeff0","lines":true,"logarithmic":false}},"legendPosition":"bottom"}},"sort":null,"type":"chart"}', '2020-12-02 15:06:49');`,
		`INSERT INTO reports.modules (id, name, text, updated) VALUES (38, 'Sales by unit, all time', '{"desc":"Created on 12/02/2020","data":"places","joins":[{"sid":"places","tid":"sales","tf":"place_id","id":"sales/place_id//places"}],"query":"","columns":[{"id":"sum.sales.total","name":"Sum Total","type":"number","meta":{},"model":""},{"id":"places.name","name":"Name","type":"text","ref":"","mid":"places","model":"Places","meta":{}}],"group":{"by":[{"id":"places.id"},{"id":"places.name"}],"columns":[{"op":"sum","id":"sales.total","name":"Sum Total","type":"number"}]},"meta":{"value":"sum.sales.total","label":"places.name","color":"sum.sales.total"},"sort":null,"type":"heatmap"}', '2020-12-02 14:32:12');`,
		`INSERT INTO reports.modules (id, name, text, updated) VALUES (39, 'All purchase by Sandra', '{"desc":"Created on 12/02/2020","data":"persons","joins":[{"sid":"persons","tid":"sales","tf":"customer_id","id":"sales/customer_id//persons"}],"query":"{\\"glue\\":\\"and\\",\\"rules\\":[{\\"field\\":\\"persons.name\\",\\"type\\":\\"text\\",\\"condition\\":{\\"filter\\":\\"Sandra\\",\\"type\\":\\"contains\\"},\\"includes\\":[]}]}","columns":[{"id":"persons.name","name":"Name","type":"text","ref":"","mid":"persons","model":"Persons","meta":{},"width":134},{"id":"persons.age","name":"Age","type":"number","ref":"","mid":"persons","model":"Persons","meta":{},"width":66},{"id":"sales.place_id","name":"Place","type":"reference","ref":"places","mid":"sales","model":"Sales","meta":{}},{"id":"sales.product_id","name":"Product","type":"reference","ref":"products","mid":"sales","model":"Sales","meta":{},"width":106},{"id":"sales.total","name":"Total","type":"number","ref":"","mid":"sales","model":"Sales","meta":{},"width":90},{"id":"sales.type","name":"Payment","type":"picklist","ref":"regions","mid":"sales","model":"Sales","meta":{},"width":104},{"id":"sales.saledate","name":"Sale Date","type":"date","ref":"","mid":"sales","model":"Sales","meta":{}}],"group":null,"meta":{"freeze":0},"sort":[{"id":"persons.name","mod":"asc"},{"id":"sales.saledate","mod":"asc"}],"type":"table"}', '2020-12-02 15:04:49');`,
	}

	for _, q := range queries {
		_, err := db.Exec(q)
		if err != nil {
			return err
		}
	}

	return nil
}