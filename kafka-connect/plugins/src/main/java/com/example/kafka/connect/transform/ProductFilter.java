package com.example.kafka.connect.transform;

import org.apache.kafka.connect.connector.ConnectRecord;
import org.apache.kafka.connect.data.Schema;
import org.apache.kafka.connect.data.Struct;
import org.apache.kafka.connect.header.Headers;
import org.apache.kafka.connect.transforms.Transformation;
import org.apache.kafka.common.Configurable;
import org.apache.kafka.common.config.ConfigDef;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;

import java.io.File;
import java.io.IOException;
import java.util.*;

public class ProductFilter<R extends ConnectRecord<R>> implements Transformation<R>, Configurable {

    private static final String ALLOWED_PRODUCTS_CONFIG = "allowed.products.file";
    private static final ConfigDef CONFIG_DEF = new ConfigDef()
        .define(ALLOWED_PRODUCTS_CONFIG, ConfigDef.Type.STRING, ConfigDef.Importance.HIGH,
            "Path to JSON file containing allowed product IDs");

    private Set<String> allowedProductIds;
    private ObjectMapper objectMapper;
    private String configFilePath;

    @Override
    public void configure(Map<String, ?> configs) {
        configFilePath = (String) configs.get(ALLOWED_PRODUCTS_CONFIG);
        objectMapper = new ObjectMapper();
        loadAllowedProducts();
    }

    @Override
    public R apply(R record) {
        if (record.value() == null) {
            return record;
        }

        try {
            JsonNode jsonNode;
            
            if (record.value() instanceof String) {
                jsonNode = objectMapper.readTree((String) record.value());
            } else if (record.value() instanceof byte[]) {
                jsonNode = objectMapper.readTree((byte[]) record.value());
            } else if (record.value() instanceof Map) {
                jsonNode = objectMapper.valueToTree(record.value());
            } else {
                return record;
            }

            String productId = jsonNode.has("product_id") ? jsonNode.get("product_id").asText() : null;

            if (productId == null) {
                return null;
            }

            if (!allowedProductIds.contains(productId)) {
                return null;
            }

            return record;

        } catch (Exception e) {
            return record;
        }
    }

    @Override
    public void close() {
    }

    @Override
    public ConfigDef config() {
        return CONFIG_DEF;
    }

    private void loadAllowedProducts() {
        try {
            File file = new File(configFilePath);
            if (!file.exists()) {
                allowedProductIds = new HashSet<>();
                return;
            }

            JsonNode root = objectMapper.readTree(file);
            JsonNode productIdsNode = root.get("allowed_product_ids");
            
            allowedProductIds = new HashSet<>();
            if (productIdsNode != null && productIdsNode.isArray()) {
                for (JsonNode idNode : productIdsNode) {
                    allowedProductIds.add(idNode.asText());
                }
            }
        } catch (IOException e) {
            allowedProductIds = new HashSet<>();
        }
    }

    public void reloadAllowedProducts() {
        loadAllowedProducts();
    }
}
