package mrgnet

import (
  "Fmt"
  "bytes"
  "crypto/rand"
  "encoding/gob"
  "io/ioutil"
  "math/big"
  "net/http"
  "net/url"
)

type NetId int64
type GameKey string

const Host_url = "http://localhost:8080"

func DoAction(name string, input, output interface{}) error {
  buf := bytes.NewBuffer(nil)
  err := gob.NewEncoder(buf).Encode(input)
  if err != nil {
    return err
  }
  host_url := fmt.Sprintf("%s/%s", Host_url, name)
  r, err := http.PostForm(host_url, url.Values{"data": []string{string(buf.Bytes())}})
  if err != nil {
    return err
  }
  data, err := ioutil.ReadAll(r.Body)
  if err != nil {
    panic(err.Error())
    return nil
  }
  dec := gob.NewDecoder(bytes.NewBuffer(data))
  return dec.Decode(output)
}

// Creates a random id that will be unique among all other engines with high
// probability.
func RandomId() NetId {
  b := big.NewInt(1 << 62)
  v, err := rand.Int(rand.Reader, b)
  if err != nil {
    // uh-oh
    panic(err)
  }
  return NetId(v.Int64())
}

type User struct {
  Id   NetId
  Name string
}

type UpdateUserRequest User
type UpdateUserResponse struct {
  User
  Err string
}

type NewGameRequest struct {
  Id NetId
}

type NewGameResponse struct {
  Err      string
  Name     string
  Game_key GameKey
}

type ListGamesRequest struct {
  Id        NetId
  Unstarted bool
}

type ListGamesResponse struct {
  Err       string
  Games     []Game
  Game_keys []GameKey
}

// Updates an active game by appending a Playback, or updating the last
// playback, with either new State or new Execs
type UpdateGameRequest struct {
  Id        NetId
  Game_key  GameKey
  Round     int
  Intruders bool

  // Exactly one of the following two should be set
  State []byte
  Execs []byte
}

type UpdateGameResponse struct {
  Err string
}

type JoinGameRequest struct {
  Id       NetId
  Game_key GameKey
}

type JoinGameResponse struct {
  Err        string
  Successful bool
}

type StatusRequest struct {
  Id       NetId
  Game_key GameKey
}

type StatusResponse struct {
  Err  string
  Game *Game
}

type Game struct {
  Name string

  Denizens_name  string
  Denizens_id    NetId
  Intruders_name string
  Intruders_id   NetId

  State [][]byte
  Execs [][]byte

  // If this is non-zero then the game is over and the winner is the player
  // whose NetId matches this value
  Winner NetId
}
