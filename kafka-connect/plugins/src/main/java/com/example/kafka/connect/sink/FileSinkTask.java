package com.example.kafka.connect.sink;

import org.apache.kafka.connect.sink.SinkRecord;
import org.apache.kafka.connect.sink.SinkTask;

import java.io.*;
import java.util.*;

public class FileSinkTask extends SinkTask {

    private String fileName;

    @Override
    public String version() {
        return "1.0.0";
    }

    @Override
    public void start(Map<String, String> props) {
        fileName = props.get("file");
        if (fileName == null) {
            fileName = "/data/output/filtered-products.txt";
        }
    }

    @Override
    public void put(Collection<SinkRecord> records) {
        if (records == null || records.isEmpty()) {
            return;
        }

        try {
            File file = new File(fileName);
            file.getParentFile().mkdirs();
            
            try (FileWriter writer = new FileWriter(file, true)) {
                for (SinkRecord record : records) {
                    if (record.value() != null) {
                        writer.write(record.value().toString());
                        writer.write("\n");
                    }
                }
            }
        } catch (IOException e) {
            e.printStackTrace();
        }
    }

    @Override
    public void stop() {
    }
}
