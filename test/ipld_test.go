package test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/codec/dagjson"
	"github.com/ipld/go-ipld-prime/must"
	"github.com/ipld/go-ipld-prime/node/basicnode"
	"github.com/ipld/go-ipld-prime/node/bindnode"
	"github.com/ipld/go-ipld-prime/schema"
)

func TestCreateData(t *testing.T) {

	np := basicnode.Prototype.Any // Pick a prototype: this is how we decide what implementation will store the in-memory data.
	nb := np.NewBuilder()         // Create a builder.
	ma, _ := nb.BeginMap(2)       // Begin assembling a map.
	ma.AssembleKey().AssignString("hey")
	ma.AssembleValue().AssignString("it works!")
	ma.AssembleKey().AssignString("yes")
	ma.AssembleValue().AssignBool(true)

	ma.Finish()     // Call 'Finish' on the map assembly to let it know no more data is coming.
	n := nb.Build() // Call 'Build' to get the resulting Node.  (It's immutable!)

	buf := []byte{}
	b := bytes.NewBuffer(buf)

	dagjson.Encode(n, b)

	fmt.Println(b.String())

}

func TestStructWithSchema(t *testing.T) {
	type Person struct {
		Name    string
		Age     int
		Friends []string
	}

	ts, err := ipld.LoadSchemaBytes([]byte(`
			type Person struct {
				name    String
				age     Int
				friends [String]
			} representation tuple
		`))
	if err != nil {
		panic(err)
	}
	schemaType := ts.TypeByName("Person")
	person := &Person{Name: "Alice", Age: 34, Friends: []string{"Bob"}}
	node := bindnode.Wrap(person, schemaType)

	fmt.Printf("%#v\n", person)

	buf := []byte{}
	b := bytes.NewBuffer(buf)

	dagjson.Encode(node.Representation(), b)

	fmt.Println(b.String())

}

func TestMarshel(t *testing.T) {
	type Foobar struct {
		Foo string
		Bar string
	}
	encoded, err := ipld.Marshal(dagjson.Encode, &Foobar{"wow", "whee"}, nil)
	fmt.Printf("error: %v\n", err)
	fmt.Printf("data: %s\n", string(encoded))

}

func TestUnmarshal(t *testing.T) {
	serial := strings.NewReader(`{"hey":"it works!","yes": true}`)

	np := basicnode.Prototype.Any // Pick a stle for the in-memory data.
	nb := np.NewBuilder()         // Create a builder.
	dagjson.Decode(nb, serial)    // Hand the builder to decoding -- decoding will fill it in!
	n := nb.Build()               // Call 'Build' to get the resulting Node.  (It's immutable!)

	fmt.Printf("the data decoded was a %s kind\n", n.Kind())
	fmt.Printf("the length of the node is %d\n", n.Length())

}

func TestUnmarshalWithSchema(t *testing.T) {
	typesys := schema.MustTypeSystem(
		schema.SpawnStruct("Foobar",
			[]schema.StructField{
				schema.SpawnStructField("foo", "String", false, false),
				schema.SpawnStructField("bar", "String", false, false),
			},
			schema.SpawnStructRepresentationMap(nil),
		),
		schema.SpawnString("String"),
	)

	type Foobar struct {
		Foo string
		Bar string
	}
	serial := []byte(`{"foo":"wow","bar":"whee"}`)
	foobar := Foobar{}
	n, err := ipld.Unmarshal(serial, dagjson.Decode, &foobar, typesys.TypeByName("Foobar"))
	fmt.Printf("error: %v\n", err)
	fmt.Printf("go struct: %v\n", foobar)
	fmt.Printf("node kind and length: %s, %d\n", n.Kind(), n.Length())
	fmt.Printf("node lookup 'foo': %q\n", must.String(must.Node(n.LookupByString("foo"))))

}
