/*
 * Copyright (c) 2022 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

import {GenericTgzArchiver} from "./generic_tgz_archiver";
import {err, Result} from "neverthrow";

export class WebTgzArchiver implements GenericTgzArchiver{

    public async createTgzByteArray(pathToArchive: string): Promise<Result<Uint8Array, Error>> {
        return err(new Error("Sending compressed archives over the Web API is not implemented yet." +
            "Please use the Node.js API instead."))
    }
}