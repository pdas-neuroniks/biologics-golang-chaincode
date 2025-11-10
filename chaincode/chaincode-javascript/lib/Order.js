"use strict";

const { Contract } = require("fabric-contract-api");

const OrderStatus = {
  DRAFT: "draft",
  THERAPY_REQUESTED: "therapy_requested",
  THERAPY_CONFIRMED: "therapy_confirmed",
  MATERIAL_READY_FOR_PICKUP: "material_ready_for_pickup",
  SHIPPED_TO_MANUFACTURER: "shipped_to_manufacturer",
  DELIVERED_TO_MANUFACTURER: "delivered_to_manufacturer",
  PROCESSING_STARTED: "processing_started",
  READY_FOR_DISPATCH: "ready_for_dispatch",
  SHIPPED_TO_HOSPITAL: "shipped_to_hospital",
  DELIVERED_TO_HOSPITAL: "delivered_to_hospital",
  THERAPY_CANCELLED: "therapy_cancelled",
  COMPLETED: "completed",
  ENTERED_IN_ERROR: "entered-in-error",
};

class OrderContract extends Contract {
  isValidStatus(status) {
    return Object.values(OrderStatus).includes(status);
  }

  /**
   * Create a new order.
   */
  async createOrder(ctx, inputData) {
    try {
      const orderData = JSON.parse(inputData);
      const order = {
        ...orderData,
        orderId: orderData.orderId,
        therapyType: orderData.therapyType,
        manufacturerId: orderData.manufacturerId,
        hospitalId: orderData.hospitalId,
        logisticsId: orderData.logisticsId,
        slotId: orderData.slotId,
        currentStatus: orderData.status,
        statusHistory: [
          {
            status: orderData.status,
            updatedBy: orderData.createdBy,
            timestamp: orderData.statusTimestamp,
          },
        ],
        createdAt: orderData.createdAt,
        ccnCode: orderData.ccnCode,
        cmsCertNumber: orderData.cmsCertNumber,
      };
      await ctx.stub.putState(orderData.orderId, Buffer.from(JSON.stringify(order)));
      return JSON.stringify(order);
    } catch (error) {
      throw new Error(`Failed to create order: ${error.message}`);
    }
  }

  async updateOrderStatus(
    ctx,
    statusUpdateData,
  ) {
    let orderUpdate = JSON.parse(statusUpdateData);
    let { orderId } = orderUpdate;
    if (!this.isValidStatus(orderUpdate.status)) {
      throw new Error(
        `Invalid status '${
          orderUpdate.status
        }'. Must be one of: ${Object.values(OrderStatus).join(", ")}`
      );
    }

    const orderBytes = await ctx.stub.getState(orderId);
    if (!orderBytes || orderBytes.length === 0) {
      throw new Error(`Order with ID ${orderId} does not exist`);
    }

    const order = JSON.parse(orderBytes.toString());

    order.statusHistory.push({
      ...orderUpdate,
      status: orderUpdate.status,
      updatedBy: orderUpdate.updatedBy,
      timestamp: orderUpdate.timestamp,
    });

    order.currentStatus = orderUpdate.status;

    await ctx.stub.putState(orderId, Buffer.from(JSON.stringify(order)));
    return JSON.stringify(order);
  }

  async getOrder(ctx, orderId) {
    const orderBytes = await ctx.stub.getState(orderId);
    if (!orderBytes || orderBytes.length === 0) {
      throw new Error(`Order with ID ${orderId} does not exist`);
    }
    return orderBytes.toString();
  }

  async orderExists(ctx, orderId) {
    const data = await ctx.stub.getState(orderId);
    return data && data.length > 0;
  }

  async getOrderHistory(ctx, orderId) {
    const iterator = await ctx.stub.getHistoryForKey(orderId);
    const allResults = [];

    while (true) {
      const res = await iterator.next();
      if (res.value && res.value.value.toString()) {
        const jsonRes = {
          txId: res.value.txId,
          timestamp: res.value.timestamp,
          isDelete: res.value.isDelete,
          value: JSON.parse(res.value.value.toString("utf8")),
        };
        allResults.push(jsonRes);
      }
      if (res.done) {
        await iterator.close();
        break;
      }
    }
    return JSON.stringify(allResults);
  }

  async getAllOrdersWithPagination(ctx, pageSize, bookmark, sortField, sortOrder) {
    // Default sorting if not provided
    const defaultSortField = sortField || "createdAt";
    const defaultSortOrder = sortOrder || "desc";
    
    const query = {
      selector: {
        orderId: { $exists: true },
      },
      sort: [
        { [defaultSortField]: defaultSortOrder }
      ],
    };

    const queryString = JSON.stringify(query);
    const pageSizeInt = parseInt(pageSize, 10);
    const { iterator, metadata } = await ctx.stub.getQueryResultWithPagination(
      queryString,
      pageSizeInt,
      bookmark || ""
    );

    const allResults = [];

    while (true) {
      const res = await iterator.next();
      if (res.value && res.value.value.toString()) {
        const jsonRes = {
          Key: res.value.key,
          Record: JSON.parse(res.value.value.toString("utf8")),
        };
        allResults.push(jsonRes);
      }
      if (res.done) {
        await iterator.close();
        break;
      }
    }

    return JSON.stringify({
      data: allResults,
      metadata: {
        fetchedRecordsCount: metadata.fetchedRecordsCount,
        bookmark: metadata.bookmark,
      },
    });
  }
}

module.exports.OrderContract = OrderContract;
