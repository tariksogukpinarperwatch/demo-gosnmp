package main

import (
	"context"
	"demo-gosnmp/config"
	"demo-gosnmp/database"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gosnmp/gosnmp"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type OID struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UserID      primitive.ObjectID `bson:"user_id" json:"user_id"`
	Description string             `bson:"description" json:"description"`
	Value       string             `bson:"value" json:"value"`
}

type SNMPConnection struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	IPAddress       string             `bson:"ip_address" json:"ip_address"`
	Port            int                `bson:"port" json:"port"`
	CommunityString string             `bson:"community_string" json:"community_string"`
	Version         string             `bson:"version" json:"version"`
	OID             string             `bson:"oid" json:"oid"`
}

func main() {

	config.LoadConfig()

	if err := database.Connect(); err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := database.Mg.Client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()

	app := fiber.New()

	app.Post("/snmp/request", snmpRequest)
	app.Post("/snmp/oids", addOID) // OID ekleme
	app.Post("/snmp/query-all", queryAllOIDs)

	port := os.Getenv("PORT")
	if port == "" {
		port = "6060" // Default port
	}

	log.Fatal(app.Listen(":" + port))
}

func addOID(c *fiber.Ctx) error {
	var oid SNMPConnection
	if err := c.BodyParser(&oid); err != nil {
		log.Println("Error parsing body:", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request format"})
	}

	collection := database.Mg.Client.Database("your_database").Collection("oids")
	oid.ID = primitive.NewObjectID()
	if _, err := collection.InsertOne(context.TODO(), oid); err != nil {
		log.Println("Error inserting OID:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to add OID"})
	}

	return c.Status(fiber.StatusCreated).JSON(oid)
}

func queryAllOIDs(c *fiber.Ctx) error {
	var req SNMPConnection
	if err := c.BodyParser(&req); err != nil {
		log.Println("Error parsing body:", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request format"})
	}

	collection := database.Mg.Client.Database("your_database").Collection("oids")

	var oids []SNMPConnection
	cursor, err := collection.Find(context.TODO(), bson.M{})
	if err != nil {
		log.Println("Error fetching OIDs:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch OIDs"})
	}
	defer cursor.Close(context.TODO())
	if err := cursor.All(context.TODO(), &oids); err != nil {
		log.Println("Error decoding OIDs:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to decode OIDs"})
	}

	// SNMP ayarları
	g := &gosnmp.GoSNMP{
		Target:    req.IPAddress,
		Port:      uint16(req.Port),
		Community: req.CommunityString,
		Version:   gosnmp.Version2c,
		Timeout:   10 * time.Second,
		Retries:   3,
	}

	// Cihaza bağlan
	if err := g.Connect(); err != nil {
		log.Println("Error connecting to SNMP device:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to connect to SNMP device"})
	}
	defer g.Conn.Close()

	// Veritabanından alınan tüm OID’lerle sorgu yap
	var oidList []string
	for _, oid := range oids {
		oidList = append(oidList, oid.OID)
	}

	result, err := g.Get(oidList)
	if err != nil {
		log.Println("Error getting OIDs:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get OIDs"})
	}

	// SNMP sonuçlarını işlem ve JSON yanıtı oluştur
	snmpResults := make(map[string]string)
	for _, variable := range result.Variables {
		var valueString string
		switch variable.Type {
		case gosnmp.OctetString:
			valueString = string(variable.Value.([]byte))
		case gosnmp.Integer:
			valueString = fmt.Sprintf("%d", variable.Value.(int))
		default:
			valueString = fmt.Sprintf("%v", variable.Value)
		}
		snmpResults[variable.Name] = valueString
	}

	return c.JSON(snmpResults)
}

func snmpRequest(c *fiber.Ctx) error {
	var conn SNMPConnection
	if err := c.BodyParser(&conn); err != nil {
		log.Println("Error parsing body:", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request format"})
	}

	g := &gosnmp.GoSNMP{
		Target:    conn.IPAddress,
		Port:      uint16(conn.Port),
		Community: conn.CommunityString,
		Version:   gosnmp.Version2c,
		Timeout:   10 * time.Second,
		Retries:   3,
	}

	if err := g.Connect(); err != nil {
		log.Println("Error connecting to SNMP device:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to connect to SNMP device"})
	}
	defer g.Conn.Close()

	if conn.OID != "" {
		result, err := g.Get([]string{conn.OID})
		if err != nil {
			log.Println("Error getting OID:", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get OID"})
		}

		snmpResults := make(map[string]string)
		for _, variable := range result.Variables {
			var valueString string
			switch variable.Type {
			case gosnmp.OctetString:
				valueString = string(variable.Value.([]byte))
			case gosnmp.Integer:
				valueString = fmt.Sprintf("%d", variable.Value.(int))
			default:
				valueString = fmt.Sprintf("%v", variable.Value)
			}
			snmpResults[variable.Name] = valueString
		}
		return c.JSON(snmpResults)
	}

	return c.JSON(fiber.Map{"message": "No OID provided."})
}
