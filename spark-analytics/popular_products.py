import json
import sys
from pyspark.sql import SparkSession
from pyspark.sql.functions import col, count, desc

def main():
    spark = SparkSession.builder \
        .appName("Analytics recommendations") \
        .config("spark.hadoop.dfs.replication", "1") \
        .getOrCreate()

    spark.sparkContext.setLogLevel("WARN")

    hdfs_path = sys.argv[1] if len(sys.argv) > 1 else "hdfs://namenode:9000/kafka_data"
    requests_path = hdfs_path + "/requests"
    output_path = hdfs_path + "/recommendations/part-00000"

    print(f"Reading data from: {requests_path}")

    try:
        df = spark.read.json(requests_path)
    except Exception as e:
        print(f"No data found: {e}")
        spark.stop()
        return

    print("Processing requests...")

    requests_df = df.select(
        col("product_name").alias("search_term")
    ).filter(col("search_term").isNotNull())

    popular_products = requests_df.groupBy("search_term") \
        .agg(count("*").alias("search_count")) \
        .orderBy(desc("search_count")) \
        .limit(10)

    print("Top popular products:")
    popular_products.show()

    output_data = popular_products.collect()

    popular_list = []
    for row in output_data:
        term = row.search_term if hasattr(row, 'search_term') else str(row[0])
        cnt = row['search_count'] if hasattr(row, 'search_count') else int(row[1])
        popular_list.append({
            "product_name": term,
            "search_count": cnt
        })

    result_json = json.dumps(popular_list, ensure_ascii=False, indent=2)

    spark.sparkContext.parallelize([result_json]).saveAsTextFile(output_path)

    print(f"Recommendations saved to: {output_path}")

    spark.stop()

if __name__ == "__main__":
    main()