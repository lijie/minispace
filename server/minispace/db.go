// +build useredis

package minispace

//import "github.com/fzzy/radix/redis"
import "github.com/garyburd/redigo/redis"

const (
	DB_GET = 1
	DB_SET = 2
)

type UserDb struct {
	Name      string
	Pass      string
	LoginTime int
	RegTime   int
	BestScore int
}

type ScoreBoard struct {
	c *redis.Client
	f *Fifo
}

type query struct {
	op        int
	key       string
	data      []byte
	replyFunc func(relyData []byte)
	err       error
}

func (s *ScoreBoard) nextQuery() *query {
	q, err := s.F.Get()
	return q
}

func (s *ScoreBoard) dbRoutine() {
	var reply interface{}
	var err error
	var q *query

	for s.enable {
		reply, err = s.c.Receive()
		q = s.nextQuery()
		if q == nil {
			continue
		}
		q.replyFunc(err, reply)
	}
}

func (s *ScoreBoard) Dial(url string) error {
}

func (s *ScoreBoard) Close() {
	s.c.Close()
}

func (s *ScoreBoard) getRedis(key string) ([]byte, error) {
	reply := s.c.Cmd("get", key)
	return reply.Bytes()
}

func (s *ScoreBoard) setRedis(key string, data []byte) error {
	reply := s.c.Cmd("set", key, data)
	return nil
}

func (s *ScoreBoard) sendQuery(q *query) error {
	var mutex sync.Mutex

	mutex.Lock()
	q.replyFunc = func(err error, replyData []byte) {
		q.data = replyData
		q.err = err
		mutex.Unlock()
	}

	// send to query queue
	err := s.f.Put(q)
	if err != nil {
		fmt.Printf("err %v\n", err)
		return err
	}

	// wait reply
	mutex.Lock()
	if q.err != nil {
		return q.err
	}

	return nil
}

func (s *ScoreBoard) Get(key string) ([]byte, error) {
	var q query
	var err error

	err = s.c.Send("get", key)
}

func (s *ScoreBoard) Set(key string, data []byte) error {
	var q query

	q.op = DB_SET
	q.key = key
	q.data = data

	err = s.sendQuery(&q)
	if err != nil {
		return err
	}

	return nil
}

func (s *ScoreBoard) newPool(server string) {
	return &redis.Pool{
		MaxIdle:     16,
		IdleTimeout: 300 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", server)
			if err != nil {
				return nil, err
			}
			return c, nil
		},
	}
}
