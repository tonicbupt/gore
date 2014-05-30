// Copyright 2014 keimoon. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Gore is a full feature Redis client for Go:
  - Convenient command building and reply parsing
  - Pipeline, multi-exec, LUA scripting
  - Pubsub
  - Connection pool
  - Redis sentinel
  - Client implementation of sharding

Connections

Gore only supports TCP connection for Redis. The connection is thread-safe and can be auto-repaired
with or without sentinel.

  conn, err := gore.Dial("localhost:6379", 10 * time.Duration) //Connect to redis server at localhost:6379
  if err != nil {
    return
  }
  defer conn.Close()

Command

Redis command is built with NewCommand

  gore.NewCommand("SET", "kirisame", "marisa") // SET kirisame marisa
  gore.NewCommand("ZADD", "magician", 1337, "alice") // ZADD magician 1337 alice
  gore.NewCommand("HSET", "sdm", "sakuya", 99) // HSET smd sakuya 99

In the last command, the value stored by redis will be the string "99", not the integer 99.
  Integer and float values are converted to string using strconv
  Boolean values are convert to "1" and "0"
  Nil values are stored as zero length string
  Other types are converted to string using standard fmt.Sprint

To efficiently store integer, you can use gore.FixInt or gore.VarInt

Compact integer

Gore supports compacting integer to reduce memory used by redis. There are 2 ways of compacting integer:
  gore.FixInt stores an integer as a fixed 8 bytes []byte.
  gore.VarInt encodes an integer with variable length []byte.

  gore.NewCommand("SET", "fixint", gore.FixInt(1337)) // Set fixint as an 8 bytes []byte
  gore.NewCommand("SET", "varint", gore.VarInt(1337)) // varint only takes 3 bytes

Reply

A redis reply is return when the command is run on a connection

  rep, err := gore.NewCommand("GET", "kirisame").Run(conn)

Parsing the reply is straightforward:

  s, _ := rep.String()  // Return string value if reply is simple string (status) or bulk string
  b, _ := rep.Bytes()   // Return a byte array
  x, _ := rep.Integer() // Return integer value if reply type is integer (INCR, DEL)
  e, _ := rep.Error()   // Return error message if reply type is error
  a, _ := rep.Array()   // Return reply list if reply type is array (MGET, ZRANGE)

Reply converting

Reply support convenient methods to convert to other types

  x, _ := rep.Int()    // Convert string value to int64. This method is different from rep.Integer()
  f, _ := rep.Float()  // Convert string value to float64
  t, _ := rep.Bool()   // Convert string value to boolean, where "1" is true and "0" is false
  x, _ := rep.FixInt() // Convert string value to FixInt
  x, _ := rep.VarInt() // Convert string value to VarInt

To convert an array reply to a slice, you can use Slice method:

  s := []int
  err := rep.Slice(&s) // Convert an array reply to a slice of integer

The following slice element types are supported:
  - integer (int, int64)
  - float (float64)
  - string and []byte
  - FixInt and VarInt
  - *gore.Pair for converting map data from HGETALL or ZRANGE WITHSCORES

Pipeline

Gore supports pipelining using gore.Pipeline:

  p := gore.NewPipeline()
  p.Add(gore.NewCommand("SET", "kirisame", "marisa"))
  p.Add(gore.NewCommand("SET", "alice", "margatroid"))
  replies, _ := p.Run(conn)
  for _, r := range replies {
      // Deal with individual reply here
  }

Script

Script can be set from a string or read from a file, and can be executed over
a connection. Gore makes sure to use EVALSHA before using EVAL to save bandwidth.

  s := gore.NewScript()
  s.SetBody("return redis.call('SET', KEYS[1], ARGV[1])")
  rep, err := s.Execute(conn, 1, "kirisame", "marisa")

Script can be loaded from a file:

  s := gore.NewScript()
  s.ReadFromFile("scripts/set.lua")
  rep, err := s.Execute(conn, 1, "kirisame", "marisa")

Script map

If your application use a lot of script files, you can manage them through ScriptMap

  gore.LoadScripts("scripts", ".*\\.lua") // Load all .lua file from scripts folder
  s := gore.GetScripts("set.lua") // Get script from set.lua file
  rep, err := s.Execute(conn, 1, "kirisame", "marisa") // And execute

Pubsub

Publish message to a channel is easy, you can use gore.Command to issue a PUBLISH
over a connection, or use gore.Publish method:

  gore.Publish(conn, "touhou", "Hello!")

To handle subscriptions, you should allocate a dedicated connection and assign it
to gore.Subscriptions:

  subs := gore.NewSubscriptions(conn)
  subs.Subscribe("test")
  subs.PSubscribe("tou*")

To receive messages, the subcriber should spawn a new goroutine and use
Subscriptions Message channel:

  go func() {
      for message := range subs.Message() {
          if message == nil {
               break
          }
          fmt.Println("Got message from %s, originate from %s: %s", message.Channel, message.OriginalChannel, message.Message)
      }
  }()
*/
package gore