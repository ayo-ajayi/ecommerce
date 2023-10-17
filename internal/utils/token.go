package utils

import (
	"errors"
	"time"

	"github.com/ayo-ajayi/ecommerce/internal/database"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TokenDetails struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	AccessUuid   string `json:"-"`
	RefreshUuid  string `json:"-"`
	AtExpires    int64  `json:"at_expires"`
	RtExpires    int64  `json:"rt_expires"`
}

type AccessDetails struct {
	AccessUuid string
	UserId     primitive.ObjectID
}

type RefreshDetails struct {
	RefreshUuid string
	UserId      primitive.ObjectID
}

type TokenManager struct {
	accessTokensecretKey        string
	refreshTokensecretKey       string
	accessTokenValidityInMins   int64
	refreshTokenValidityInHours int64
	redisClient                 *redis.Client
}

func NewTokenManager(accessTokensecretKey, refreshTokensecretKey string, accessTokenValidityInMins int64, refreshTokenValidityInHours int64, redisClient *redis.Client) *TokenManager {
	return &TokenManager{
		accessTokensecretKey,
		refreshTokensecretKey,
		accessTokenValidityInMins,
		refreshTokenValidityInHours,
		redisClient,
	}
}

func (tu *TokenManager) GenerateToken(userId primitive.ObjectID) (*TokenDetails, error) {
	td := &TokenDetails{}
	td.AtExpires = time.Now().Add(time.Minute * time.Duration(tu.accessTokenValidityInMins)).Unix()
	td.AccessUuid = uuid.New().String()
	td.RtExpires = time.Now().Add(time.Hour * time.Duration(tu.refreshTokenValidityInHours)).Unix()
	td.RefreshUuid = uuid.New().String()

	var err error
	td.AccessToken, err = createToken("access_uuid", userId, td.AccessUuid, td.AtExpires, tu.accessTokensecretKey)
	if err != nil {
		return nil, err
	}
	td.RefreshToken, err = createToken("refresh_uuid", userId, td.RefreshUuid, td.RtExpires, tu.refreshTokensecretKey)
	if err != nil {
		return nil, err
	}
	if td.RefreshToken == "" || td.AccessToken == "" {
		return nil, errors.New("token generation failed")
	}
	return td, nil
}

func createToken(uuidType string, userId primitive.ObjectID, uuid string, expires int64, secretKey string) (string, error) {
	claims := jwt.MapClaims{}
	claims["authorized"] = true
	claims[uuidType] = uuid
	claims["user_id"] = userId
	claims["exp"] = expires
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}

func (tu *TokenManager) SaveToken(userId primitive.ObjectID, td *TokenDetails) error {
	ctx, cancel := database.DBReqContext(5)
	defer cancel()
	at := time.Unix(td.AtExpires, 0)
	rt := time.Unix(td.RtExpires, 0)
	now := time.Now()
	txPipeline := tu.redisClient.TxPipeline()
	txPipeline.Set(ctx, td.AccessUuid, userId.Hex(), at.Sub(now)).Err()
	txPipeline.Set(ctx, td.RefreshUuid, userId.Hex(), rt.Sub(now)).Err()
	_, err := txPipeline.Exec(ctx)
	return err
}

func (tu *TokenManager) FindToken(uuid string) (string, error) {
	ctx, cancel := database.DBReqContext(5)
	defer cancel()
	exists, err := tu.redisClient.Exists(ctx, uuid).Result()
	if err != nil {
		return "", err
	}
	if exists == 0 {
		return "", errors.New("token not found")
	}
	return tu.redisClient.Get(ctx, uuid).Result()
}

func (tu *TokenManager) DeleteToken(uuid string) error {
	ctx, cancel := database.DBReqContext(5)
	defer cancel()
	exists, err := tu.redisClient.Exists(ctx, uuid).Result()
	if err != nil {
		return err
	}
	if exists == 0 {
		return errors.New("token not found")
	}
	return tu.redisClient.Del(ctx, uuid).Err()
}

func ExtractTokenDetails(token *jwt.Token) (*AccessDetails, error) {
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid access token")
	}
	accessUuid, ok := claims["access_uuid"].(string)
	if !ok || accessUuid == "" {
		return nil, errors.New("invalid access token claims")
	}
	userId, ok := claims["user_id"].(string)
	if !ok || userId == "" {
		return nil, errors.New("invalid access token claims")
	}
	userID, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		return nil, err
	}
	return &AccessDetails{
		AccessUuid: accessUuid,
		UserId:     userID,
	}, nil
}

func ValidateToken(token string, secretKey string) (*jwt.Token, error) {
	return jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(secretKey), nil
	})
}

func (tu *TokenManager) IdentifyUser(refreshToken string) (*RefreshDetails, error) {
	token, err := ValidateToken(refreshToken, tu.refreshTokensecretKey)
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid refresh token")
	}
	refreshUuid, ok := claims["refresh_uuid"].(string)
	if !ok || refreshUuid == "" {
		return nil, errors.New("invalid refresh token claims")
	}
	userId, ok := claims["user_id"].(string)
	if !ok || userId == "" {
		return nil, errors.New("invalid refresh token claims")
	}
	userID, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		return nil, err
	}
	return &RefreshDetails{
		RefreshUuid: refreshUuid,
		UserId:      userID,
	}, nil
}
