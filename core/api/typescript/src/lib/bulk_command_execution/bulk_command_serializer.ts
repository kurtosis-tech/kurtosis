import { V0BulkCommands } from "./v0_bulk_command_api/v0_bulk_commands";
import { SchemaVersion } from "./bulk_command_schema_version";
import { ok, err, Result } from "neverthrow";

const LATEST_SCHEMA_VERSION: SchemaVersion = SchemaVersion.V0;

class VersionedBulkCommandsDocument {
    private readonly schemaVersion: SchemaVersion;
    
    constructor(schemaVersion: SchemaVersion) {
        this.schemaVersion = schemaVersion;
    }
}

class SerializableBulkCommandsDocument extends VersionedBulkCommandsDocument {
    private readonly body: any;
    
    constructor(schemaVersion: SchemaVersion, body: V0BulkCommands) {
        super(schemaVersion);
        this.body = body;
    }
}


class BulkCommandSerializer {

    constructor (){}

    public serialize(bulkCommands: V0BulkCommands): Result<Uint8Array | string, Error> {
        const toSerialize: SerializableBulkCommandsDocument = new SerializableBulkCommandsDocument(LATEST_SCHEMA_VERSION, bulkCommands);
        
        let bytes: string;
        try {
            bytes = JSON.stringify(toSerialize);
        } catch (jsonErr) {
            // Sadly, we have to do this because there's no great way to enforce the caught thing being an error
            // See: https://stackoverflow.com/questions/30469261/checking-for-typeof-error-in-js
            if (jsonErr && jsonErr.stack && jsonErr.message) {
                return err(jsonErr as Error);
            }
            return err(new Error("Stringify-ing SerializableBulkCommandsDocument object threw an exception, but " +
                "it's not an Error so we can't report any more information than this"));
        }

        return ok(bytes);
    }
}
