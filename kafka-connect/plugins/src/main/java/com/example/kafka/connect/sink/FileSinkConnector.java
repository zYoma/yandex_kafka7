package com.example.kafka.connect.sink;

import org.apache.kafka.connect.connector.Task;
import org.apache.kafka.connect.source.SourceConnector;
import org.apache.kafka.connect.sink.SinkConnector;
import org.apache.kafka.common.config.ConfigDef;
import org.apache.kafka.common.utils.Utils;

import java.util.*;

public class FileSinkConnector extends SinkConnector {

    private static final String FILE_CONFIG = "file";
    private static final String FILE_DOC = "File output path";
    private static final String FILE_DEFAULT = "/data/output/filtered-products.txt";

    private String fileName;

    @Override
    public void start(Map<String, String> props) {
        fileName = props.get(FILE_CONFIG);
        if (fileName == null) {
            fileName = FILE_DEFAULT;
        }
    }

    @Override
    public Class<? extends Task> taskClass() {
        return FileSinkTask.class;
    }

    @Override
    public List<Map<String, String>> taskConfigs(int maxTasks) {
        ArrayList<Map<String, String>> configs = new ArrayList<>();
        Map<String, String> config = new HashMap<>();
        config.put(FILE_CONFIG, fileName);
        configs.add(config);
        return configs;
    }

    @Override
    public void stop() {
    }

    @Override
    public ConfigDef config() {
        ConfigDef config = new ConfigDef();
        config.define(FILE_CONFIG, ConfigDef.Type.STRING, FILE_DEFAULT, ConfigDef.Importance.HIGH, FILE_DOC);
        return config;
    }

    @Override
    public String version() {
        return "1.0.0";
    }
}
