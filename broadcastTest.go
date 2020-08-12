// package broadcast
package main

// func TestBroadcast(t *testing.T) {
// 	b := NewBroadcaster()

// 	r := b.Listen()

// 	b.Write("hello")

// 	if r.Read().(string) != "hello" {
// 		t.Fatal("error string")
// 	}

// 	r1 := b.Listen()

// 	b.Write(123)

// 	if r.Read().(int) != 123 {
// 		t.Fatal("error int")
// 	}

// 	if r1.Read().(int) != 123 {
// 		t.Fatal("error int")
// 	}

// 	b.Write(nil)

// 	if r.Read() != nil {
// 		t.Fatal("error nit")
// 	}

// 	if r1.Read() != nil {
// 		t.Fatal("error nit")
// 	}
// }
