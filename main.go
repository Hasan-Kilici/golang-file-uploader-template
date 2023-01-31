package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Username  string
	Password  string
	CreatedAt time.Time
}

type Image struct {
	ID  primitive.ObjectID `bson:"_id,omitempty"`
	Src string             `bson:"src"`
}

func main() {

	clientOptions := options.Client().ApplyURI("mongodb+srv://codearmy:code2009@cluster0.juzwe.mongodb.net")

	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("MongoDB bağlantısı başarılı!")

	usercollection := client.Database("ccc").Collection("users")
	imagecollection := client.Database("ccc").Collection("images")

	r := gin.Default()
	r.LoadHTMLGlob("src/*.tmpl")
	r.Static("/static", "./static/")
	r.Static("/uploads", "./upload/")
	//GET
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"title": "Anasayfa",
		})
	})

	r.GET("/register", func(c *gin.Context) {
		c.HTML(http.StatusOK, "register.tmpl", gin.H{
			"title": "Kayıt ol",
		})
	})

	r.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.tmpl", gin.H{
			"title": "Giriş yap",
		})
	})

	r.GET("/photos", func(c *gin.Context) {
		c.HTML(http.StatusOK, "photos.tmpl", gin.H{
			"title": "Fotoğraflar",
		})
	})

	r.GET("/upload-photo", func(c *gin.Context) {
		token, err := c.Cookie("token")
		if err != nil {
			c.String(http.StatusOK, "Çerez bulunamadı.")
			return
		} else {
			c.HTML(http.StatusOK, "upload.tmpl", gin.H{
				"title": "Fotoğraf yükle",
			})
			fmt.Println(token)
		}
	})

	r.GET("/photo/:id", func(c *gin.Context) {
		id := c.Param("id")
		c.HTML(http.StatusOK, "image.tmpl", gin.H{
			"fotoId": id,
		})
	})

	r.GET("sign-out", func(c *gin.Context) {
		cookie := http.Cookie{
			Name:   "token",
			Value:  "",
			Path:   "/",
			MaxAge: -1,
		}
		c.Writer.Header().Set("Set-Cookie", cookie.String())
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"title": "Anasayfa",
		})
	})
	//API
	r.GET("/images", func(c *gin.Context) {
		var images []Image
		cur, err := imagecollection.Find(context.TODO(), bson.D{{}})
		if err != nil {
			log.Fatalf("Error finding images: %v", err)
		}
		for cur.Next(context.TODO()) {
			var image Image
			err := cur.Decode(&image)
			if err != nil {
				log.Fatalf("Error decoding image: %v", err)
			}
			images = append(images, image)
		}
		if err := cur.Err(); err != nil {
			log.Fatalf("Error getting next image: %v", err)
		}
		c.JSON(http.StatusOK, images)
	})

	r.GET("/images/:id", func(c *gin.Context) {
		id := c.Param("id")
		oid, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			log.Fatalf("Error converting to ObjectID: %v", err)
		}
		var image Image
		err = imagecollection.FindOne(context.TODO(), bson.M{"_id": oid}).Decode(&image)
		if err != nil {
			log.Fatalf("Error finding image: %v", err)
		}
		c.JSON(http.StatusOK, image)
	})

	//POST
	r.POST("/register", func(c *gin.Context) {
		username := c.PostForm("username")
		password := c.PostForm("password")

		user := User{
			ID:        primitive.NewObjectID(),
			Username:  username,
			Password:  password,
			CreatedAt: time.Now(),
		}
		insertResult, err := usercollection.InsertOne(context.TODO(), user)
		if err != nil {
			log.Fatal(err)
		}

		c.SetCookie("token", insertResult.InsertedID.(primitive.ObjectID).Hex(), 3600, "/", "", false, true)
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"message": fmt.Sprintf("Kayıt başarılı. Kullanıcı ID'niz: %s", insertResult.InsertedID),
		})
	})

	r.POST("/login", func(c *gin.Context) {
		username := c.PostForm("username")
		password := c.PostForm("password")

		var user User
		err := usercollection.FindOne(context.TODO(), bson.M{"username": username, "password": password}).Decode(&user)
		if err != nil {
			c.HTML(http.StatusOK, "index.tmpl", gin.H{
				"message": "Kullanıcı adı veya şifre hatalı.",
			})
			return
		}

		c.SetCookie("token", user.ID.Hex(), 3600, "/", "", false, true)
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"message": fmt.Sprintf("Giriş başarılı. Hoşgeldiniz, %s", username),
		})
	})

	r.POST("/upload", func(c *gin.Context) {
		cookie, err := c.Cookie("token")
		if err != nil {
			c.String(http.StatusOK, "Çerez bulunamadı.")
			return
		} else {

			c.String(http.StatusOK, fmt.Sprintf("Çerez değeri: %s", cookie))
			file, _ := c.FormFile("file")
			fmt.Println(file.Filename)
			dst := "upload/" + file.Filename

			image := Image{
				ID:  primitive.NewObjectID(),
				Src: dst,
			}
			c.SaveUploadedFile(file, dst)
			insertResult, err := imagecollection.InsertOne(context.TODO(), image)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println("Fotoğraf Yüklendi Yüklenen fotoğraf:", insertResult)
			c.String(http.StatusOK, fmt.Sprintf("'%s' yüklendi!", file.Filename))

		}
	})

	r.Run(":5000")
}
