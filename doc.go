// Copyright 2015 Ola Holmström. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package comet implements a framework for dealing with WebSockets.
//
// Example
//
// A broadcasting echo server:
//
//  func main() {
//  	r := gin.Default()
//  	m := comet.New()
//  	r.GET("/ws", func(c *gin.Context) {
//  		m.HandleRequest(c.Writer, c.req)
//  	})
//  	m.HandleMessage(func(s *comet.Session, msg []byte) {
//  		m.Broadcast(msg)
//  	})
//  	r.Run(":5000")
//  }
package comet
