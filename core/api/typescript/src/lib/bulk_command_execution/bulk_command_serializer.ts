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
            if(jsonErr instanceof Error){
                return err(jsonErr);
            }
            return err(new Error("Stringify-ing SerializableBulkCommandsDocument object threw an exception, but " +
                "it's not an Error so we can't report any more information than this"));
        }

        return ok(bytes);
    }
}
