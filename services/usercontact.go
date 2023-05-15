package services

import (
	"fmt"
	"net"
	"time"

	"github.com/lib/pq"
)

type UserContact struct {
	AccountID AccountId
	Ip        net.IP
	Port      int32
	Time      time.Time
}

func (u UserContact) GetURL() string {
	return fmt.Sprintf("%s:%d", u.Ip.String(), u.Port)
}

func UpdateUserContact(appCtx AppCtx, accountId AccountId, ip net.IP, port int32) error {
	log := appCtx.Log("UpdateUserContact")
	log.Printf("Received: %v/%v/%v", accountId, ip, port)

	stmt := `insert into user_contacts (id,ip,port,timestamp) VALUES ($1,$2::INET,$3,now()) ON CONFLICT(id) DO UPDATE SET ip=$2::INET, port=$3, timestamp=now() `
	result, err := appCtx.db.Exec(stmt, accountId, ip.String(), port)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	log.Printf("Removed %d rows\n", rows)

	return nil
}

func GetUserContacts(appCtx AppCtx, accountIDs []AccountId) ([]UserContact, error) {
	log := appCtx.Log("GetUserContacts")
	log.Printf("Received: %v", accountIDs)

	//::uuid[]
	stmt := `select id, ip, port, timestamp from "user_contacts" where id = any($1::uuid[])`

	params := make([]string, len(accountIDs))
	for idx, accountId := range accountIDs {
		params[idx] = accountId.String()
	}

	rows, err := appCtx.db.Query(stmt, pq.Array(params))
	if err != nil {
		return []UserContact{}, err
	}

	contacts := make([]UserContact, 0)
	defer rows.Close()
	for rows.Next() {
		var id string
		var ip string
		var port int32
		var timestamp time.Time

		rows.Scan(&id, &ip, &port, &timestamp)
		contacts = append(contacts, UserContact{
			AccountID: NewAccountId(id),
			Ip:        net.ParseIP(ip),
			Port:      port,
			Time:      timestamp,
		})
	}

	return contacts, nil

}
