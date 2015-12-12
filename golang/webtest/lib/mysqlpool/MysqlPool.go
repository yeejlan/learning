
package mysqlpool

import(
    "../pools"
    "time"
    "github.com/ziutek/mymysql/mysql"
    _ "github.com/ziutek/mymysql/native" // Native engine
)

type MysqlPool struct{
    *pools.RoundRobin
    Host string
    User string
    DbName string
}

const connIdleTimeout = "5m" //5 minutes 

func NewMysqlPool(host string, user string, pass string, dbname string, capacity int) *MysqlPool {
    
    idleTimeout, err := time.ParseDuration(connIdleTimeout)
    if err!=nil {
        panic(err)
    }
    p := &MysqlPool{pools.NewRoundRobin(capacity, idleTimeout), host, user, dbname}
    p.open(mysqlCreator(host, user, pass, dbname))
    return p
}

func (this *MysqlPool) open(connFactory CreateMysqlFunc){
	if connFactory == nil {
		return
	}
	f := func() (pools.Resource, error) {
		c, err := connFactory()
		if err != nil {
			return nil, err
		}
		return &Mysql{c, this}, nil
	}
	this.RoundRobin.Open(f)
}

// You must call Recycle on the *Mysql once done.
func (this *MysqlPool)Get() *Mysql{
	r, err := this.RoundRobin.Get()
	if err != nil {
		panic(err)
	}
	return r.(*Mysql)
}

type CreateMysqlFunc func() (mysql.Conn, error)

func mysqlCreator(host, user, pass, dbname string) CreateMysqlFunc {
	return func() (mysql.Conn, error) {
	    db := mysql.New("tcp", "", host, user, pass, dbname)
		err := db.Connect()
        if err != nil {
            panic(err)
        }
        return db, nil
	}
}

// Cache re-exposes mysql.Conn
type Mysql struct {
	Conn mysql.Conn
	pool *MysqlPool
}

func (this *Mysql) Recycle() {
	this.pool.Put(this)
}


func (this *Mysql) IsClosed() bool {
	return this.Conn.IsConnected()
}

func (this *Mysql) Close() {
    this.Conn.Close()
}




