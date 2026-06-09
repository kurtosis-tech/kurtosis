CREATE DATABASE IF NOT EXISTS otel;

-- Database for the ethereum-package observoor eBPF profiler. observoor's
-- migrator creates its own tables on startup but expects the database to exist.
CREATE DATABASE IF NOT EXISTS observoor;

CREATE TABLE IF NOT EXISTS otel.otel_logs
(
    Timestamp                  DateTime64(9) CODEC(Delta(8), ZSTD(1)),
    TraceId                    String CODEC(ZSTD(1)),
    SpanId                     String CODEC(ZSTD(1)),
    TraceFlags                 UInt8,
    SeverityText               LowCardinality(String) CODEC(ZSTD(1)),
    SeverityNumber             UInt8,
    ServiceName                LowCardinality(String) CODEC(ZSTD(1)),
    Body                       String CODEC(ZSTD(1)),
    ResourceSchemaUrl          LowCardinality(String) CODEC(ZSTD(1)),
    ResourceAttributes         Map(LowCardinality(String), String) CODEC(ZSTD(1)),
    EnclaveName                LowCardinality(String) MATERIALIZED ResourceAttributes['kurtosis.enclave.name'],
    EnclaveUuid                LowCardinality(String) MATERIALIZED ResourceAttributes['kurtosis.enclave.uuid'],
    ScopeSchemaUrl             LowCardinality(String) CODEC(ZSTD(1)),
    ScopeName                  String CODEC(ZSTD(1)),
    ScopeVersion               LowCardinality(String) CODEC(ZSTD(1)),
    ScopeAttributes            Map(LowCardinality(String), String) CODEC(ZSTD(1)),
    LogAttributes              Map(LowCardinality(String), String) CODEC(ZSTD(1)),
    EventName                  String CODEC(ZSTD(1)),
    INDEX idx_trace_id         TraceId                       TYPE bloom_filter(0.001)    GRANULARITY 1,
    INDEX idx_span_id          SpanId                        TYPE bloom_filter(0.001)    GRANULARITY 1,
    INDEX idx_body             Body                          TYPE tokenbf_v1(8192, 3, 0) GRANULARITY 4,
    INDEX idx_res_attr_key     mapKeys(ResourceAttributes)   TYPE bloom_filter(0.01)     GRANULARITY 1,
    INDEX idx_res_attr_value   mapValues(ResourceAttributes) TYPE bloom_filter(0.01)     GRANULARITY 1,
    INDEX idx_scope_attr_key   mapKeys(ScopeAttributes)      TYPE bloom_filter(0.01)     GRANULARITY 1,
    INDEX idx_scope_attr_value mapValues(ScopeAttributes)    TYPE bloom_filter(0.01)     GRANULARITY 1,
    INDEX idx_log_attr_key     mapKeys(LogAttributes)        TYPE bloom_filter(0.01)     GRANULARITY 1,
    INDEX idx_log_attr_value   mapValues(LogAttributes)      TYPE bloom_filter(0.01)     GRANULARITY 1
)
ENGINE = ReplacingMergeTree
PARTITION BY toDate(Timestamp)
ORDER BY (
    EnclaveName,
    EnclaveUuid,
    toStartOfFiveMinutes(Timestamp),
    ServiceName,
    Timestamp,
    cityHash64(concat(Body, LogAttributes['kurtosis.line_index']))
)
TTL toDateTime(Timestamp) + INTERVAL 6 HOUR DELETE
SETTINGS index_granularity = 8192, ttl_only_drop_parts = 1;


CREATE TABLE IF NOT EXISTS otel.otel_traces
(
    Timestamp                 DateTime64(9) CODEC(Delta(8), ZSTD(1)),
    TraceId                   String CODEC(ZSTD(1)),
    SpanId                    String CODEC(ZSTD(1)),
    ParentSpanId              String CODEC(ZSTD(1)),
    TraceState                String CODEC(ZSTD(1)),
    SpanName                  LowCardinality(String) CODEC(ZSTD(1)),
    SpanKind                  LowCardinality(String) CODEC(ZSTD(1)),
    ServiceName               LowCardinality(String) CODEC(ZSTD(1)),
    ResourceAttributes        Map(LowCardinality(String), String) CODEC(ZSTD(1)),
    EnclaveName               LowCardinality(String) MATERIALIZED ResourceAttributes['kurtosis.enclave.name'],
    EnclaveUuid               LowCardinality(String) MATERIALIZED ResourceAttributes['kurtosis.enclave.uuid'],
    ScopeName                 String CODEC(ZSTD(1)),
    ScopeVersion              String CODEC(ZSTD(1)),
    SpanAttributes            Map(LowCardinality(String), String) CODEC(ZSTD(1)),
    Duration                  UInt64 CODEC(ZSTD(1)),
    StatusCode                LowCardinality(String) CODEC(ZSTD(1)),
    StatusMessage             String CODEC(ZSTD(1)),
    Events Nested (
        Timestamp  DateTime64(9),
        Name       LowCardinality(String),
        Attributes Map(LowCardinality(String), String)
    ) CODEC(ZSTD(1)),
    Links Nested (
        TraceId    String,
        SpanId     String,
        TraceState String,
        Attributes Map(LowCardinality(String), String)
    ) CODEC(ZSTD(1)),
    INDEX idx_trace_id        TraceId                       TYPE bloom_filter(0.001) GRANULARITY 1,
    INDEX idx_span_id         SpanId                        TYPE bloom_filter(0.001) GRANULARITY 1,
    INDEX idx_duration        Duration                      TYPE minmax              GRANULARITY 1,
    INDEX idx_res_attr_key    mapKeys(ResourceAttributes)   TYPE bloom_filter(0.01)  GRANULARITY 1,
    INDEX idx_res_attr_value  mapValues(ResourceAttributes) TYPE bloom_filter(0.01)  GRANULARITY 1,
    INDEX idx_span_attr_key   mapKeys(SpanAttributes)       TYPE bloom_filter(0.01)  GRANULARITY 1,
    INDEX idx_span_attr_value mapValues(SpanAttributes)     TYPE bloom_filter(0.01)  GRANULARITY 1
)
ENGINE = MergeTree
PARTITION BY toDate(Timestamp)
ORDER BY (EnclaveName, EnclaveUuid, ServiceName, SpanName, toDateTime(Timestamp))
TTL toDateTime(Timestamp) + INTERVAL 6 HOUR DELETE
SETTINGS index_granularity = 8192, ttl_only_drop_parts = 1;


CREATE TABLE IF NOT EXISTS otel.otel_traces_trace_id_ts
(
    TraceId String CODEC(ZSTD(1)),
    Start   DateTime CODEC(Delta, ZSTD(1)),
    End     DateTime CODEC(Delta, ZSTD(1)),

    INDEX idx_trace_id TraceId TYPE bloom_filter(0.01) GRANULARITY 1
)
ENGINE = MergeTree
PARTITION BY toDate(Start)
ORDER BY (TraceId, Start)
TTL Start + INTERVAL 6 HOUR DELETE
SETTINGS index_granularity = 8192, ttl_only_drop_parts = 1;


CREATE MATERIALIZED VIEW IF NOT EXISTS otel.otel_traces_trace_id_ts_mv
TO otel.otel_traces_trace_id_ts AS
SELECT
    TraceId,
    toDateTime(min(Timestamp)) AS Start,
    toDateTime(max(Timestamp)) AS End
FROM otel.otel_traces
WHERE TraceId != ''
GROUP BY TraceId;
