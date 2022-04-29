/*
 * Copyright (c) 2022 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

import {GenericTgzArchiver} from "./generic_tgz_archiver";
import {err, Result} from "neverthrow";

export class WebFileArchiver implements GenericTgzArchiver{

    public async createTgz(pathToArchive: string): Promise<Result<Uint8Array, Error>> {
        return err(new Error("Web File Archiver has not been implemented yet."))
    }
}