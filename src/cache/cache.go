package cache

import (
	"context"
	"encoding/json"
	"os"      // Adicionado para ler variáveis de ambiente
	"strconv" // Adicionado para converter string para int
	"time"

	"github.com/go-redis/redis/v8"
)

var (
	// Cliente Redis global
	Client *redis.Client
	// Contexto para operações do Redis
	Ctx = context.Background()
)

// Setup inicializa o cliente Redis
func Setup(addr, password string, db int) {
	Client = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
}

// SetupWithURL inicializa o cliente Redis a partir de uma URL
func SetupWithURL(redisURL string) error {
	options, err := redis.ParseURL(redisURL)
	if err != nil {
		return err
	}

	Client = redis.NewClient(options)
	return nil
}

// Get retorna um valor do cache
func Get(key string) (string, error) {
	return Client.Get(Ctx, key).Result()
}

// Set define um valor no cache com TTL
func Set(key string, value interface{}, ttl time.Duration) error {
	return Client.Set(Ctx, key, value, ttl).Err()
}

// Delete remove uma chave do cache
func Delete(key string) error {
	return Client.Del(Ctx, key).Err()
}

// FlushAll limpa todo o cache
func FlushAll() error {
	return Client.FlushAll(Ctx).Err()
}

// GetCached verifica se um item existe no cache e o retorna deserializado se encontrado
func GetCached(key string, result interface{}) (bool, error) {
	if Client == nil {
		return false, nil // Cache está desabilitado
	}

	cachedData, err := Client.Get(Ctx, key).Result()
	if err == redis.Nil {
		return false, nil // Chave não existe no cache
	} else if err != nil {
		return false, err // Erro de conexão ou outro problema
	}

	// Deserializar o JSON no tipo fornecido
	if err := json.Unmarshal([]byte(cachedData), result); err != nil {
		return false, err
	}

	return true, nil // Cache hit
}

// GetTTL retorna o TTL recomendado para diferentes tipos de dados
func GetTTL(dataType string) time.Duration {
	switch dataType {
	case "character":
		// Tenta ler o TTL do character do ambiente
		ttlStr := os.Getenv("CACHE_TTL_CHARACTER")
		if ttlStr != "" {
			ttl, err := strconv.Atoi(ttlStr)
			if err == nil && ttl > 0 {
				return time.Duration(ttl) * time.Second
			}
		}
		return 60 * time.Second // valor padrão se não configurado ou inválido
	case "world":
		return 10 * time.Second
	case "guild":
		return 10 * time.Second
	case "highscores":
		return 1 * time.Minute
	default:
		return 60 * time.Second
	}
}
