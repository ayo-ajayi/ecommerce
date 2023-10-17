package router

import (
	"log"
	"sync"

	"os"

	"github.com/ayo-ajayi/ecommerce/internal/app/cart"
	"github.com/ayo-ajayi/ecommerce/internal/app/category"
	"github.com/ayo-ajayi/ecommerce/internal/app/item"
	"github.com/ayo-ajayi/ecommerce/internal/app/review"
	"github.com/ayo-ajayi/ecommerce/internal/app/search"
	"github.com/ayo-ajayi/ecommerce/internal/app/user"
	"github.com/ayo-ajayi/ecommerce/internal/constants"
	"github.com/ayo-ajayi/ecommerce/internal/database"
	mw "github.com/ayo-ajayi/ecommerce/internal/middleware"
	"github.com/ayo-ajayi/ecommerce/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	cors "github.com/rs/cors/wrapper/gin"
)

func NewRouter() *gin.Engine {
	redisUri := os.Getenv("REDIS_URI")
	mongoDBUri := os.Getenv("MONGODB_URI")
	otpIssuer := os.Getenv("OTP_ISSUER")
	mongoDBName := os.Getenv("MONGODB_NAME")
	emailOfSender := os.Getenv("EMAIL_OF_SENDER")
	emailApiKey := os.Getenv("EMAIL_API_KEY")
	emailSenderName := os.Getenv("EMAIL_SENDER_NAME")
	accessTokenSecretKey := os.Getenv("ACCESS_TOKEN_SECRET_KEY")
	refreshTokenSecretKey := os.Getenv("REFRESH_TOKEN_SECRET_KEY")
	cloudinaryURI := os.Getenv("CLOUDINARY_URI")

	if redisUri == "" ||
		mongoDBUri == "" ||
		otpIssuer == "" ||
		mongoDBName == "" ||
		emailOfSender == "" ||
		emailApiKey == "" ||
		emailSenderName == "" ||
		accessTokenSecretKey == "" ||
		refreshTokenSecretKey == "" || cloudinaryURI == "" {
		log.Fatal("environment variables not set")
	}

	accessTokenValidityInMins := constants.AccessTokenValidityInMins
	refreshTokenValidityInHours := constants.RefreshTokenValidityInHours
	signUpOtpValidityInSecs := constants.SignUpOtpValidityInSecs
	forgotPasswordOtpValidityInSecs := constants.ForgotPasswordOtpValidityInSecs
	redisDBValue := constants.RedisDBValue

	client, err := database.NewMongoDBClient(mongoDBUri)
	if err != nil {
		log.Fatal(err.Error())
	}
	userCollection := database.NewMongoDBCollection(client, mongoDBName, "users")
	otpCollection := database.NewMongoDBCollection(client, mongoDBName, "otps")
	categoryCollection := database.NewMongoDBCollection(client, mongoDBName, "categories")
	itemCollection := database.NewMongoDBCollection(client, mongoDBName, "items")
	reviewCollection := database.NewMongoDBCollection(client, mongoDBName, "reviews")
	cartCollection := database.NewMongoDBCollection(client, mongoDBName, "carts")

	otpManager := utils.NewOTPManager(otpCollection, otpIssuer, signUpOtpValidityInSecs, forgotPasswordOtpValidityInSecs)

	emailManager := utils.NewEmailManager(emailOfSender, emailSenderName, emailApiKey)

	redisClient := redis.NewClient(&redis.Options{Addr: redisUri, DB: redisDBValue})
	tokenManager := utils.NewTokenManager(accessTokenSecretKey, refreshTokenSecretKey, accessTokenValidityInMins, refreshTokenValidityInHours, redisClient)
	mediaCloudManager, err := utils.NewMediaCloudManager(cloudinaryURI, "ecommerce")
	if err != nil {
		log.Fatal(err.Error())
	}

	userRepo := user.NewUserRepo(userCollection)
	userService := user.NewUserService(userRepo, otpManager, emailManager, tokenManager)
	userController := user.NewUserController(userService)

	categoryRepo := category.NewCategoryRepo(categoryCollection)
	categoryService := category.NewCategoryService(categoryRepo)
	categoryController := category.NewCategoryController(categoryService, mediaCloudManager)

	itemRepo := item.NewItemRepo(itemCollection)
	itemService := item.NewItemService(itemRepo, categoryRepo, mediaCloudManager)
	itemController := item.NewItemController(itemService)

	cartRepo := cart.NewCartRepo(cartCollection)
	cartService := cart.NewCartService(cartRepo, itemRepo)
	cartController := cart.NewCartController(cartService)

	reviewRepo := review.NewReviewRepo(reviewCollection)
	reviewService := review.NewReviewService(reviewRepo, userRepo)
	reviewController := review.NewReviewController(reviewService)

	middleware := mw.NewMiddleware(accessTokenSecretKey, tokenManager, userRepo)
	searchController := search.NewSearchController(itemRepo, categoryRepo)

	ctx, cancel := database.DBReqContext(10)
	defer cancel()
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := client.Ping(ctx, nil); err != nil {
			log.Fatal(err.Error())
		}
		log.Println("mongodb connected")
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := redisClient.Ping(ctx).Err(); err != nil {
			log.Fatal(err.Error())
		}
		log.Println("redis connected")
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := search.InitSearchIndex(itemCollection, categoryCollection); err != nil {
			log.Fatal(err.Error())
		}
		if err := utils.InitOtpExpiryIndex(otpCollection); err != nil {
			log.Fatal(err.Error())
		}
	}()
	wg.Wait()
	router := gin.Default()
	router.Use(middleware.JsonMiddleware(), cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
	}))
	router.GET("/favicon.ico", func(ctx *gin.Context) { ctx.File("./favicon.ico") })
	router.NoRoute(func(ctx *gin.Context) { ctx.JSON(404, gin.H{"error": "endpoint not found"}) })
	router.GET("/healthz", func(ctx *gin.Context) { ctx.JSON(200, gin.H{"message": "ok"}) })
	router.GET("/", func(ctx *gin.Context) { ctx.JSON(200, gin.H{"message": "welcome to ecommerce"}) })
	api := router.Group("/api")
	{
		api.GET("", func(ctx *gin.Context) { ctx.JSON(200, gin.H{"message": "welcome to ecommerce api"}) })
		api.POST("/signup", userController.SignUp)
		api.POST("/verify", userController.VerifyUser)
		api.POST("/login", userController.Login)
		api.POST("/forgot-password", userController.SendForgotPasswordOTP)
		api.POST("/reset-password", userController.ResetPassword)
		api.POST("/refresh-token", userController.RefreshToken)
		api.POST("/resend-verification-otp", userController.ResendEmailVerificationOTP)
		api.GET("/search", searchController.Search)

	}
	all := api.Group("")
	{
		all.GET("/category/:id", categoryController.GetCategoryByID)
		all.GET("/category/slug/:slug", categoryController.GetCategoryBySlug)
		all.GET("/categories", categoryController.GetCategories)
		all.GET("/item/:id", itemController.GetItemByID)
		all.GET("/items", itemController.GetItems)
		all.GET("/reviews", reviewController.GetReviews)
		all.GET("/review/:id", reviewController.GetReview)
		all.GET("/item/slug/:slug", itemController.GetItemBySlug)

		authenticated := all.Group("", middleware.Authentication())
		{
			authenticated.GET("/profile", userController.Profile)
			authenticated.POST("/logout", userController.Logout)
			authenticated.POST("/add-address", userController.AddAddress)
			authenticated.GET("/addresses", userController.GetAddresses)
			authenticated.GET("/address/:id", userController.GetAddress)
			authenticated.PUT("/update-address/:id", userController.UpdateAddress)
			authenticated.DELETE("/remove-address/:id", userController.RemoveAddress)
			customer := authenticated.Group("/customer", middleware.Authorization([]user.Role{user.Customer}))
			{
				customer.POST("/post-review", reviewController.PostReview)
				customer.PUT("/update-cart", cartController.UpdateCart)
				customer.GET("/cart", cartController.GetCart)
			}
			vendor := authenticated.Group("/vendor", middleware.Authorization([]user.Role{user.Vendor}))
			{
				vendor.POST("/create-item", itemController.CreateItem)
				vendor.PUT("/update-item/:id", itemController.UpdateItem)
				vendor.DELETE("/delete-item/:id", itemController.DeleteItem)
				vendor.GET("/items", itemController.GetVendorItems)

			}
			admin := authenticated.Group("/admin", middleware.Authorization([]user.Role{user.Admin}))
			{
				admin.POST("/create-category", categoryController.CreateCategory)
				admin.PUT("/update-category/:id", categoryController.UpdateCategory)
				admin.DELETE("/delete-category/:id", categoryController.DeleteCategory)
				admin.GET("/users", userController.GetUsers)
			}
		}
	}

	return router
}
