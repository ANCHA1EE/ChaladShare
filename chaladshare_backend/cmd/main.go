package main

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"chaladshare_backend/internal/config"
	"chaladshare_backend/internal/connect"
	"chaladshare_backend/internal/connectdb"
	"chaladshare_backend/internal/middleware"

	AuthHandler "chaladshare_backend/internal/auth/handlers"
	AuthRepo "chaladshare_backend/internal/auth/repository"
	AuthService "chaladshare_backend/internal/auth/service"

	FileHandler "chaladshare_backend/internal/files/handlers"
	FileRepo "chaladshare_backend/internal/files/repository"
	FileService "chaladshare_backend/internal/files/service"

	PostHandler "chaladshare_backend/internal/posts/handlers"
	PostRepo "chaladshare_backend/internal/posts/repository"
	PostService "chaladshare_backend/internal/posts/service"

	UserHandler "chaladshare_backend/internal/users/handlers"
	UserRepo "chaladshare_backend/internal/users/repository"
	UserService "chaladshare_backend/internal/users/service"

	FriendsHandler "chaladshare_backend/internal/friends/handlers"
	FriendsRepo "chaladshare_backend/internal/friends/repository"
	FriendsService "chaladshare_backend/internal/friends/service"

	FeatureHandler "chaladshare_backend/internal/docfeatures/handlers"
	FeatureRepo "chaladshare_backend/internal/docfeatures/repository"
	FeatureService "chaladshare_backend/internal/docfeatures/service"

	RecommendHandler "chaladshare_backend/internal/recommend/handlers"
	RecommendRepo "chaladshare_backend/internal/recommend/repository"
	RecommendService "chaladshare_backend/internal/recommend/service"
)

func TimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func parseAllowOrigins(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return []string{
			"http://localhost:3000",
			"http://127.0.0.1:3000",
		}
	}

	// รองรับได้ทั้ง "a,b,c" หรือ "a"
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func main() {
	_ = godotenv.Load()

	for _, k := range []string{"SUPABASE_URL", "SUPABASE_SERVICE_ROLE_KEY", "SUPABASE_STORAGE_BUCKET"} {
		if os.Getenv(k) == "" {
			log.Println("WARNING:", k, "is empty")
		}
	}

	// config
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// add test colab
	if os.Getenv("COLAB_URL") == "" {
		log.Println("WARNING: COLAB_URL is empty")
	} else {
		log.Println("COLAB_URL =", os.Getenv("COLAB_URL"))
	}

	// connect DB
	db, err := connectdb.NewPostgresDatabase(cfg.GetConnectionString())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// cookie secure flag (Railway/Vercel ต้อง true)
	secureCookie := strings.ToLower(os.Getenv("COOKIE_SECURE")) == "true"
	accessCookieName := os.Getenv("ACCESS_COOKIE_NAME")
	if accessCookieName == "" {
		accessCookieName = "access_token"
	}

	refreshCookieName := os.Getenv("REFRESH_COOKIE_NAME")
	if refreshCookieName == "" {
		refreshCookieName = "refresh_token"
	}

	// auth
	authRepository := AuthRepo.NewAuthRepository(db.GetDB())
	authService := AuthService.NewAuthService(authRepository, []byte(cfg.JWTSecret), cfg.TokenTTLMinutes)
	authHandler := AuthHandler.NewAuthHandler(authService, accessCookieName, refreshCookieName, secureCookie)

	// friends
	friendsRepo := FriendsRepo.NewFriendRepository(db.GetDB())
	friendsService := FriendsService.NewFriendService(friendsRepo)
	friendsHandler := FriendsHandler.NewFriendHandler(friendsService)

	// AI client (Colab/ngrok)
	aiClient, err := connect.NewFromEnv()
	if err != nil {
		log.Printf("WARNING: cannot init AI client: %v", err)
		aiClient = nil
	}

	featureRepository := FeatureRepo.NewFeatureRepo(db.GetDB())
	featureService := FeatureService.NewFeatureService(featureRepository, aiClient)
	featureHandler := FeatureHandler.NewFeatureHandler(featureService)

	// file
	fileRepository := FileRepo.NewFileRepository(db.GetDB())
	fileService := FileService.NewFileService(fileRepository, featureService)
	fileHandler := FileHandler.NewFileHandler(fileService)

	if aiClient != nil {
		// ใช้ interface assertion เพื่อไม่พังแม้ยังไม่ได้เพิ่ม method ใน interface
		if b, ok := any(featureService).(interface{ BootstrapAutoClustering() }); ok {
			b.BootstrapAutoClustering()
			log.Println("[AUTO-CLUSTER] bootstrap started")
		} else {
			log.Println("[AUTO-CLUSTER] BootstrapAutoClustering() not found - add it to FeatureService interface")
		}
	} else {
		log.Println("[AUTO-CLUSTER] skip bootstrap: aiClient is nil")
	}

	// recommend
	recommendRepo := RecommendRepo.NewRecommendRepo(db.GetDB())
	recommendService := RecommendService.NewRecommendService(recommendRepo, aiClient)
	recommendHandler := RecommendHandler.NewRecommendHandler(recommendService, recommendRepo)

	// post like save
	postRepository := PostRepo.NewPostRepository(db.GetDB())
	postService := PostService.NewPostService(postRepository, friendsService, fileService)

	likeRepository := PostRepo.NewLikeRepository(db.GetDB())
	likeService := PostService.NewLikeService(likeRepository)
	likeHandler := PostHandler.NewLikeHandler(likeService, recommendService)

	saveRepository := PostRepo.NewSaveRepository(db.GetDB())
	saveService := PostService.NewSaveService(saveRepository)

	postHandler := PostHandler.NewPostHandler(postService, saveService)

	// user
	userRepository := UserRepo.NewUserRepository(db.GetDB())
	userService := UserService.NewUserService(userRepository)
	userHandler := UserHandler.NewUserHandler(userService, postService, friendsService)

	go func() {
		for {
			time.Sleep(10 * time.Second)
			if err := db.Ping(); err != nil {
				log.Printf("Database connection lost: %v", err)
				if reconnErr := db.Reconnect(cfg.GetConnectionString()); reconnErr != nil {
					log.Printf("Failed to reconnect: %v", reconnErr)
				} else {
					log.Printf("Successfully reconnected to the database")
				}
			}
		}
	}()

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "ChaladShare backend",
		})
	})

	// allow origins (รองรับหลายโดเมน)
	// ใช้ cfg.AllowOrigin เดิมได้ แต่แนะนำให้ทำให้มันเป็น csv ได้ เช่น:
	// ALLOW_ORIGIN="http://localhost:3000,https://xxx.vercel.app"
	allowOrigins := parseAllowOrigins(os.Getenv("ALLOW_ORIGIN"))

	r.Use(cors.New(cors.Config{
		AllowOrigins:     allowOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.Use(TimeoutMiddleware(180 * time.Second))

	r.MaxMultipartMemory = 100 << 20
	uploadDir := os.Getenv("UPLOAD_DIR")
	if uploadDir == "" {
		uploadDir = "/tmp/uploads"
	}
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Printf("WARNING: cannot create upload dir (%s): %v", uploadDir, err)
	}
	r.Static("/uploads", uploadDir)

	// r.Static("/uploads", "./uploads")

	// 	if err := os.MkdirAll("uploads", 0755); err != nil {
	// 		log.Fatalf("cannot create uploads dir: %v", err)
	// 	}

	r.GET("/health", func(c *gin.Context) {
		if err := connectdb.CheckDBConnection(db.GetDB()); err != nil {
			c.JSON(503, gin.H{"detail": "Database connection failed"})
			return
		}
		c.JSON(200, gin.H{"status": "healthy", "database": "connected"})
	})

	v1 := r.Group("/api/v1")

	// login register
	authRoutes := v1.Group("/auth")
	{
		authRoutes.POST("/register", authHandler.Register)
		authRoutes.POST("/login", authHandler.Login)
		authRoutes.POST("/logout", authHandler.Logout)
		authRoutes.POST("/refresh", authHandler.Refresh)

		authRoutes.POST("/forgot-password", authHandler.ForgotPassword)
		authRoutes.POST("/forgot-password/verify-otp", authHandler.VerifyForgotPasswordOTP)
		authRoutes.POST("/reset-password", authHandler.ResetPassword)

		authRoutes.GET("/users", authHandler.GetAllUsers)
		authRoutes.POST("/register/request-otp", authHandler.RequestRegisterOTP)
		authRoutes.POST("/register/confirm-otp", authHandler.ConfirmVerifyEmailOTP)
	}
	// Protected (ต้องมี JWT)
	protected := v1.Group("/")
	protected.Use(middleware.JWT([]byte(cfg.JWTSecret), accessCookieName))
	{
		posts := protected.Group("/posts")
		{
			posts.GET("", postHandler.GetAllPosts)
			posts.GET("/:id", postHandler.GetPostByID)

			posts.POST("", postHandler.CreatePost)
			posts.PUT("/:id", postHandler.UpdatePost)
			posts.DELETE("/:id", postHandler.DeletePost)

			posts.POST("/:id/like", likeHandler.ToggleLike)
			posts.POST("/:id/save", postHandler.ToggleSave)
			posts.GET("/save", postHandler.GetSavedPosts)
			posts.GET("/popular", postHandler.GetPopularPosts)
			posts.GET("/search", postHandler.SearchPosts)

		}

		files := protected.Group("/files")
		{
			files.POST("/doc", fileHandler.UploadFile)
			files.GET("/user/:id", fileHandler.GetFilesByUserID)
			files.GET("/:document_id/summary", fileHandler.GetSummaryByDocumentID)
			files.DELETE("/:document_id", fileHandler.DeleteFile)

			files.POST("/cover", fileHandler.UploadCover)
			files.POST("/avatar", fileHandler.UploadAvatar)
		}

		profile := protected.Group("/profile")
		{
			profile.GET("", userHandler.GetOwnProfile)
			profile.PUT("", userHandler.UpdateOwnProfile)
			profile.GET("/:id", userHandler.GetViewedUserProfile)
			profile.POST("/change-password", userHandler.ChangePassword)

		}

		social := protected.Group("/social")
		{
			social.POST("/follow", friendsHandler.FollowUser)
			social.DELETE("/follow/:id", friendsHandler.UnfollowUser)

			social.GET("/friends/:id", friendsHandler.ListFriends)
			social.GET("/followers/:id", friendsHandler.ListFollowers)
			social.GET("/following/:id", friendsHandler.ListFollowing)

			social.GET("/stats/:id", friendsHandler.GetStats)

			social.POST("/requests", friendsHandler.SendFriendRequest)
			social.GET("/requests/incoming", friendsHandler.ListIncomingRequests)
			social.GET("/requests/outgoing", friendsHandler.ListOutgoingRequests)
			social.POST("/requests/:id/accept", friendsHandler.AcceptFriendRequest)
			social.POST("/requests/:id/decline", friendsHandler.DeclineFriendRequest)
			social.DELETE("/requests/:id", friendsHandler.CancelFriendRequest)

			social.DELETE("/friends/:id", friendsHandler.Unfriend)
			social.GET("/addfriends", friendsHandler.SearchAddFriend)

		}

		docfeatures := protected.Group("/features")
		{
			docfeatures.GET("/vectors", featureHandler.GetVectors)
			docfeatures.POST("/clusters/batch_update", featureHandler.BatchUpdateClusters)
			docfeatures.POST("/clusters/run", featureHandler.RunClustering)
		}

		recommend := protected.Group("/recommend")
		{
			recommend.GET("", recommendHandler.GetRecommend)
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = cfg.AppPort
	}
	if port == "" {
		port = "8080"
	}

	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
