package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SNMPConnection struct {
	ID                primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	IPAddress         string             `bson:"ip_address" json:"ip_address"`
	Port              int                `bson:"port" json:"port"`
	CommunityString   string             `bson:"community_string" json:"community_string"`
	User              string             `bson:"user" json:"user"`
	Version           string             `bson:"version" json:"version"`
	LastConnected     primitive.DateTime `bson:"last_connected" json:"last_connected"`
	TargetInformation string             `bson:"target_information" json:"target_information"`
	Events            []string           `bson:"events" json:"events"`
	OID               string             `bson:"oid" json:"oid"`
}
