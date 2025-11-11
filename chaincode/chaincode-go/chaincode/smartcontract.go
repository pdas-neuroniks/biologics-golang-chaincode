package chaincode

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// Order Status constants (using const group)
const (
	DRAFT                     = "draft"
	THERAPY_REQUESTED         = "therapy_requested"
	THERAPY_CONFIRMED         = "therapy_confirmed"
	MATERIAL_READY_FOR_PICKUP = "material_ready_for_pickup"
	SHIPPED_TO_MANUFACTURER   = "shipped_to_manufacturer"
	DELIVERED_TO_MANUFACTURER = "delivered_to_manufacturer"
	PROCESSING_STARTED        = "processing_started"
	READY_FOR_DISPATCH        = "ready_for_dispatch"
	SHIPPED_TO_HOSPITAL       = "shipped_to_hospital"
	DELIVERED_TO_HOSPITAL     = "delivered_to_hospital"
	THERAPY_CANCELLED         = "therapy_cancelled"
	COMPLETED                 = "completed"
	ENTERED_IN_ERROR          = "entered-in-error"
)

// SmartContract provides functions for managing an Asset
type SmartContract struct {
	contractapi.Contract
}

// =========================================================================================
// DATA STRUCTURES
// =========================================================================================

// StatusHistoryEntry mirrors the structure of an entry in the statusHistory array
type StatusHistoryEntry struct {
	Status    string `json:"status"`
	UpdatedBy string `json:"updatedBy"`
	Timestamp string `json:"timestamp"` // Using string for consistency with original JS
}

// Order defines the structure for an asset
type Order struct {
	OrderID        string               `json:"orderId"`
	TherapyType    string               `json:"therapyType"`
	ManufacturerID string               `json:"manufacturerId"`
	HospitalID     string               `json:"hospitalId"`
	LogisticsID    string               `json:"logisticsId"`
	SlotID         string               `json:"slotId"`
	CurrentStatus  string               `json:"currentStatus"`
	StatusHistory  []StatusHistoryEntry `json:"statusHistory"`
	CreatedAt      string               `json:"createdAt"` // Using string for consistency with original JS
	CCNCode        string               `json:"ccnCode"`
	CMSCertNumber  string               `json:"cmsCertNumber"`
}

// StatusUpdatePayload is used for unmarshalling the input to updateOrderStatus
type StatusUpdatePayload struct {
	OrderID   string `json:"orderId"`
	Status    string `json:"status"`
	UpdatedBy string `json:"updatedBy"`
	Timestamp string `json:"timestamp"`
}

// HistoryQueryResult is used to format the result from GetOrderHistory
type HistoryQueryResult struct {
	TxID      string `json:"txId"`
	Timestamp string `json:"timestamp"`
	IsDelete  bool   `json:"isDelete"`
	Value     *Order `json:"value,omitempty"` // omitempty handles deleted assets
}

// OrderQueryResult is used to format the result for a single order in pagination
type OrderQueryResult struct {
	Key    string `json:"Key"`
	Record *Order `json:"Record"`
}

// PaginationMetadata is for the metadata part of the pagination result
type PaginationMetadata struct {
	FetchedRecordsCount int32  `json:"fetchedRecordsCount"`
	Bookmark            string `json:"bookmark"`
}

// PaginatedQueryResponse packages the data and metadata
type PaginatedQueryResponse struct {
	Data     []OrderQueryResult `json:"data"`
	Metadata PaginationMetadata `json:"metadata"`
}

// =========================================================================================
// HELPER FUNCTIONS
// =========================================================================================

// isValidStatus checks if a given status string is one of the valid OrderStatus constants
func isValidStatus(status string) bool {
	validStatuses := map[string]bool{
		DRAFT:                     true,
		THERAPY_REQUESTED:         true,
		THERAPY_CONFIRMED:         true,
		MATERIAL_READY_FOR_PICKUP: true,
		SHIPPED_TO_MANUFACTURER:   true,
		DELIVERED_TO_MANUFACTURER: true,
		PROCESSING_STARTED:        true,
		READY_FOR_DISPATCH:        true,
		SHIPPED_TO_HOSPITAL:       true,
		DELIVERED_TO_HOSPITAL:     true,
		THERAPY_CANCELLED:         true,
		COMPLETED:                 true,
		ENTERED_IN_ERROR:          true,
	}
	_, ok := validStatuses[status]
	return ok
}

// InitLedger adds a base set of assets to the ledger
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	assets := []Order{
		{OrderID: "orderid001", TherapyType: "blue"},
	}

	for _, asset := range assets {
		assetJSON, err := json.Marshal(asset)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(asset.OrderID, assetJSON)
		if err != nil {
			return fmt.Errorf("failed to put to world state. %v", err)
		}
	}

	return nil
}

// CreateOrder creates a new order and saves it to the ledger.
func (c *SmartContract) CreateOrder(ctx contractapi.TransactionContextInterface, inputData string) (*Order, error) {
	var orderData Order

	// 1. Unmarshal input string into the Order struct
	err := json.Unmarshal([]byte(inputData), &orderData)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal input data: %v", err)
	}

	// 2. Validate essential fields
	if orderData.OrderID == "" || orderData.CurrentStatus == "" {
		return nil, fmt.Errorf("orderId and status must not be empty")
	}

	// 3. Construct the initial order object (mirroring JS logic)
	order := Order{
		OrderID:        orderData.OrderID,
		TherapyType:    orderData.TherapyType,
		ManufacturerID: orderData.ManufacturerID,
		HospitalID:     orderData.HospitalID,
		LogisticsID:    orderData.LogisticsID,
		SlotID:         orderData.SlotID,
		CurrentStatus:  orderData.CurrentStatus,
		StatusHistory: []StatusHistoryEntry{
			{
				Status:    orderData.CurrentStatus,
				UpdatedBy: orderData.StatusHistory[0].UpdatedBy, // Assuming the first entry provides this
				Timestamp: orderData.StatusHistory[0].Timestamp, // Assuming the first entry provides this
			},
		},
		CreatedAt:     orderData.CreatedAt,
		CCNCode:       orderData.CCNCode,
		CMSCertNumber: orderData.CMSCertNumber,
	}

	// 4. Marshal the Go struct back into JSON bytes
	orderJSON, err := json.Marshal(order)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal order data: %v", err)
	}

	// 5. Put the state to the ledger
	err = ctx.GetStub().PutState(order.OrderID, orderJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to put state: %v", err)
	}

	return &order, nil
}

// GetOrder retrieves an order by ID.
func (c *SmartContract) GetOrder(ctx contractapi.TransactionContextInterface, orderID string) (*Order, error) {
	orderBytes, err := ctx.GetStub().GetState(orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if orderBytes == nil {
		return nil, fmt.Errorf("order with ID %s does not exist", orderID)
	}

	var order Order
	err = json.Unmarshal(orderBytes, &order)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal order data: %v", err)
	}

	return &order, nil
}

// UpdateOrderStatus updates the current status of an existing order.
func (c *SmartContract) UpdateOrderStatus(ctx contractapi.TransactionContextInterface, statusUpdateData string) (*Order, error) {
	var updatePayload StatusUpdatePayload

	// 1. Unmarshal the update payload
	err := json.Unmarshal([]byte(statusUpdateData), &updatePayload)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal status update data: %v", err)
	}

	orderID := updatePayload.OrderID

	// 2. Validate the new status
	if !isValidStatus(updatePayload.Status) {
		return nil, fmt.Errorf("invalid status '%s'. Must be one of: %s", updatePayload.Status,
			fmt.Sprintf("%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s", DRAFT, THERAPY_REQUESTED, THERAPY_CONFIRMED, MATERIAL_READY_FOR_PICKUP, SHIPPED_TO_MANUFACTURER, DELIVERED_TO_MANUFACTURER, PROCESSING_STARTED, READY_FOR_DISPATCH, SHIPPED_TO_HOSPITAL, DELIVERED_TO_HOSPITAL, THERAPY_CANCELLED, COMPLETED, ENTERED_IN_ERROR))
	}

	// 3. Get the existing order from the ledger
	orderBytes, err := ctx.GetStub().GetState(orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if orderBytes == nil {
		return nil, fmt.Errorf("order with ID %s does not exist", orderID)
	}

	var order Order
	err = json.Unmarshal(orderBytes, &order)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal order data: %v", err)
	}

	// 4. Append the new status to the history
	newHistoryEntry := StatusHistoryEntry{
		Status:    updatePayload.Status,
		UpdatedBy: updatePayload.UpdatedBy,
		Timestamp: updatePayload.Timestamp,
	}
	order.StatusHistory = append(order.StatusHistory, newHistoryEntry)

	// 5. Update the current status
	order.CurrentStatus = updatePayload.Status

	// 6. Marshal and put the updated state
	updatedOrderJSON, err := json.Marshal(order)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal updated order data: %v", err)
	}

	err = ctx.GetStub().PutState(orderID, updatedOrderJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to update state: %v", err)
	}

	return &order, nil
}

// OrderExists checks if an order exists in the ledger.
func (c *SmartContract) OrderExists(ctx contractapi.TransactionContextInterface, orderID string) (bool, error) {
	orderBytes, err := ctx.GetStub().GetState(orderID)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return orderBytes != nil, nil
}

// GetAllOrders returns all orders found in world state
func (s *SmartContract) GetAllOrders(ctx contractapi.TransactionContextInterface) ([]*Order, error) {
	// range query with empty string for startKey and endKey does an
	// open-ended query of all orders in the chaincode namespace.
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var assets []*Order
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var asset Order
		err = json.Unmarshal(queryResponse.Value, &asset)
		if err != nil {
			return nil, err
		}
		assets = append(assets, &asset)
	}

	return assets, nil
}

// GetOrderHistory retrieves the history of changes for a specific order key.
func (c *SmartContract) GetOrderHistory(ctx contractapi.TransactionContextInterface, orderID string) ([]HistoryQueryResult, error) {
	resultsIterator, err := ctx.GetStub().GetHistoryForKey(orderID)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var history []HistoryQueryResult
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var order *Order
		if len(response.Value) > 0 {
			order = new(Order)
			if err := json.Unmarshal(response.Value, order); err != nil {
				return nil, fmt.Errorf("failed to unmarshal history value: %v", err)
			}
		}

		// Convert timestamp to time.RFC3339 format for consistency
		timestamp := time.Unix(response.Timestamp.Seconds, int64(response.Timestamp.Nanos)).Format(time.RFC3339)

		history = append(history, HistoryQueryResult{
			TxID:      response.TxId,
			Timestamp: timestamp,
			IsDelete:  response.IsDelete,
			Value:     order,
		})
	}

	return history, nil
}

// GetAllOrdersWithPagination retrieves all orders using a rich query with pagination and sorting.
// Note: In Go, arguments are typically passed directly, not as a single JSON string, but for direct
// translation compatibility and flexibility, we will accept strings and convert as needed.
func (c *SmartContract) GetAllOrdersWithPagination(ctx contractapi.TransactionContextInterface, pageSizeStr string, bookmark, sortField, sortOrder string) (*PaginatedQueryResponse, error) {

	// Parse pageSize
	pageSize, err := strconv.ParseInt(pageSizeStr, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("pageSize must be a valid integer: %v", err)
	}

	// Default sorting logic
	defaultSortField := sortField
	if defaultSortField == "" {
		defaultSortField = "createdAt"
	}
	defaultSortOrder := sortOrder
	if defaultSortOrder == "" {
		defaultSortOrder = "desc"
	}

	// Construct the rich query
	query := map[string]interface{}{
		"selector": map[string]interface{}{
			"orderId": map[string]string{
				"$exists": "true",
			},
		},
		"sort": []map[string]string{
			{defaultSortField: defaultSortOrder},
		},
	}

	fmt.Print("Query, ", query)

	queryStringBytes, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query: %v", err)
	}
	queryString := string(queryStringBytes)

	fmt.Print("Query String, ", queryString)

	// Execute the query with pagination
	resultsIterator, metadata, err := ctx.GetStub().GetQueryResultWithPagination(
		queryString,
		int32(pageSize),
		bookmark,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get query result with pagination: %v", err)
	}
	defer resultsIterator.Close()

	fmt.Print("Results Iterator: ", resultsIterator)

	var results []OrderQueryResult
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var order Order
		if err := json.Unmarshal(response.Value, &order); err != nil {
			return nil, fmt.Errorf("failed to unmarshal record: %v", err)
		}

		results = append(results, OrderQueryResult{
			Key:    response.Key,
			Record: &order,
		})
	}

	// Package the response
	paginatedResponse := &PaginatedQueryResponse{
		Data: results,
		Metadata: PaginationMetadata{
			FetchedRecordsCount: metadata.FetchedRecordsCount,
			Bookmark:            metadata.Bookmark,
		},
	}

	return paginatedResponse, nil
}
