package services

import (
	"net"
	"time"

	"github.com/hoyle1974/grapevine/common"
	"github.com/lib/pq"
)

type userContact struct {
	common.Contact
	Time time.Time
}

func UpdateUserContact(appCtx AppCtx, accountId common.AccountId, ip net.IP, port int32) error {
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

func GetUserContacts(appCtx AppCtx, accountIDs []common.AccountId) ([]common.Contact, error) {
	log := appCtx.Log("GetUserContacts")
	log.Printf("Received: %v", accountIDs)

	//::uuid[]
	stmt := `select id, ip, port from "user_contacts" where id = any($1::uuid[])`

	params := make([]string, len(accountIDs))
	for idx, accountId := range accountIDs {
		params[idx] = accountId.String()
	}

	rows, err := appCtx.db.Query(stmt, pq.Array(params))
	if err != nil {
		return []common.Contact{}, err
	}

	contacts := make([]common.Contact, 0)
	defer rows.Close()
	for rows.Next() {
		var id string
		var ip string
		var port int

		rows.Scan(&id, &ip, &port)
		contacts = append(contacts, common.NewContact(common.NewAccountId(id), net.ParseIP(ip), port))
	}

	return contacts, nil

}
