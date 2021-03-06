package rod

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/boltdb/bolt"
)

type Animal struct {
	Type string
	Name string
}

type User struct {
	Username string
	Logins   int
}

type Car struct {
	Manufacturer string
	Model        string
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func TestAll(t *testing.T) {
	dir, err := ioutil.TempDir("", "rod-")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir)

	filename := filepath.Join(dir, "rod.db")
	defer os.Remove(filename)

	// Open the database.
	db, err := bolt.Open(filename, 0666, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	t.Run("Simple Put and Get", func(t *testing.T) {
		msg := "Hello, World!"

		err := db.Update(func(tx *bolt.Tx) error {
			// put this message
			err := Put(tx, "message", "hello-world", []byte(msg))
			check(err)

			// get it back
			storedMsg, err := Get(tx, "message", "hello-world")
			check(err)
			if string(storedMsg) != msg {
				log.Fatalf("Received msg '%s' is not the same as the original '%s'", string(storedMsg), msg)
			}

			return nil
		})

		check(err)
	})

	t.Run("Simple PutString and GetString", func(t *testing.T) {
		msg := "Hello, World!"

		err := db.Update(func(tx *bolt.Tx) error {
			// put this message
			err := PutString(tx, "strings", "hello-world", msg)
			check(err)

			// get it back
			storedMsg, err := GetString(tx, "strings", "hello-world")
			check(err)
			if storedMsg != msg {
				log.Fatalf("Received msg '%s' is not the same as the original '%s'", storedMsg, msg)
			}

			return nil
		})

		check(err)
	})

	t.Run("Simple PutJson and GetJson", func(t *testing.T) {
		user := User{"chilts", 1}

		err := db.Update(func(tx *bolt.Tx) error {
			// put this message
			errPutJson := PutJson(tx, "user", "chilts", user)
			check(errPutJson)

			// get it back
			storedUser := User{}
			errGetJson := GetJson(tx, "user", "chilts", &storedUser)
			check(errGetJson)
			if storedUser.Username != user.Username {
				log.Fatalf("Received msg '%s' is not the same as the original '%s'", storedUser.Username, user.Username)
			}
			if storedUser.Logins != user.Logins {
				log.Fatalf("Received logins '%d' is not the same as the original '%d'", storedUser.Logins, user.Logins)
			}

			return nil
		})

		check(err)
	})

	t.Run("SelAll (DEPRECATED)", func(t *testing.T) {
		// Start a read-write transaction.
		if err := db.Update(func(tx *bolt.Tx) error {
			dog := Animal{"dog", "rover"}
			cat := Animal{"cat", "willow"}
			horse := Animal{"horse", "ed"}

			_ = PutJson(tx, "animal", "dog", &dog)
			_ = PutJson(tx, "animal", "cat", &cat)
			_ = PutJson(tx, "animal", "horse", &horse)

			animals := make([]*Animal, 0)
			err := SelAll(tx, "animal", func() interface{} {
				return Animal{}
			}, func(v interface{}) {
				a, notOk := v.(Animal)
				if notOk {
					t.Fatal("Thing returned from SelAll is not an Animal")
				}
				animals = append(animals, &a)
			})

			return err
		}); err != nil {
			log.Fatal(err)
		}
	})

	t.Run("Sel", func(t *testing.T) {
		// Start a read-write transaction.
		if err := db.Update(func(tx *bolt.Tx) error {
			carBucketName := "car"
			golf := Car{"Volkswagon", "Golf"}
			leaf := Car{"Nissan", "Leaf"}
			hilux := Car{"Toyota", "Hilux"}

			_ = PutJson(tx, carBucketName, "golf", &golf)
			_ = PutJson(tx, carBucketName, "leaf", &leaf)
			_ = PutJson(tx, carBucketName, "hilux", &hilux)

			var cars []Car
			err := All(tx, carBucketName, &cars)
			if err != nil {
				log.Fatal(err)
			}

			if len(cars) != 3 {
				t.Fatalf("Three cars should have been returned from All(), but instead %d were\n", len(cars))
			}

			for i, car := range cars {
				if i == 0 {
					if car.Manufacturer != golf.Manufacturer {
						t.Fatal("First car should have been a Volkswagon")
					}
					if car.Model != golf.Model {
						t.Fatal("First car should have been a Golf")
					}
				}
				if i == 1 {
					if car.Manufacturer != hilux.Manufacturer {
						t.Fatal("Second car should have been a Toyota")
					}
					if car.Model != hilux.Model {
						t.Fatal("Second car should have been a Hilux")
					}
				}
				if i == 2 {
					if car.Manufacturer != leaf.Manufacturer {
						t.Fatal("Third car should have been a Nissan")
					}
					if car.Model != leaf.Model {
						t.Fatal("Third car should have been a Leaf")
					}
				}
			}

			return err
		}); err != nil {
			log.Fatal(err)
		}
	})

	t.Run("AllKeys", func(t *testing.T) {
		// Start a read-write transaction.
		if err := db.Update(func(tx *bolt.Tx) error {
			carBucketName := "make-model"
			golf := Car{"Volkswagon", "Golf"}
			leaf := Car{"Nissan", "Leaf"}
			hilux := Car{"Toyota", "Hilux"}

			_ = PutJson(tx, carBucketName, "golf", &golf)
			_ = PutJson(tx, carBucketName, "leaf", &leaf)
			_ = PutJson(tx, carBucketName, "hilux", &hilux)

			cars, err := AllKeys(tx, carBucketName)
			if err != nil {
				log.Fatal(err)
			}

			if len(cars) != 3 {
				t.Fatalf("Three cars should have been returned from AllKeys(), but instead %d were\n", len(cars))
			}

			return nil
		}); err != nil {
			log.Fatal(err)
		}

		check(err)
	})

	t.Run("AllKeys - non-existant bucket", func(t *testing.T) {
		// view only
		if err := db.View(func(tx *bolt.Tx) error {
			carBucketName := "does-not-exist"

			cars, err := AllKeys(tx, carBucketName)
			if err != nil {
				log.Fatal(err)
			}

			if cars != nil {
				t.Fatalf("Should have been returned a nil slice due to the bucket not existing\n")
			}

			if len(cars) != 0 {
				t.Fatalf("No cars should have been returned from AllKeys(), but instead %d were\n", len(cars))
			}

			return nil
		}); err != nil {
			log.Fatal(err)
		}

		check(err)
	})

	t.Run("Delete", func(t *testing.T) {
		location := "delete"
		key := "key"
		val := "val"

		err := db.Update(func(tx *bolt.Tx) error {
			// put a new value and make sure it works
			errPut := PutString(tx, location, key, val)
			check(errPut)

			// now delete it again
			errDel1 := Del(tx, location, key)
			check(errDel1)

			// delete it again (since it now doesn't exist)
			errDel2 := Del(tx, location, key)
			check(errDel2)

			// delete a key from a location that doesn't exist
			errDel3 := Del(tx, "doesnt-exist", key)
			check(errDel3)

			return nil
		})

		check(err)
	})
}
