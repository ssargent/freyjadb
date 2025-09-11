// Example: How the lore package could use the API package
//
// This demonstrates how the API package can be reused by other parts
// of the system, such as the lore example application.
//
// package lore
//
// import (
//     "github.com/ssargent/freyjadb/pkg/api"
//     "github.com/ssargent/freyjadb/pkg/store"
// )
//
// func startLoreAPIServer(store *store.KVStore, port int, apiKey string) error {
//     config := api.ServerConfig{
//         Port:   port,
//         APIKey: apiKey,
//     }
//     return api.StartServer(store, config)
// }
//
// // Then in lore's main:
// // store := initializeLoreStore()
// // startLoreAPIServer(store, 8080, "lore-api-key")
//
// This shows how the API functionality is now cleanly separated and reusable!