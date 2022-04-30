/*
 * Copyright (c) 2022 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

import "neverthrow"
import {GenericTgzArchiver} from "./generic_tgz_archiver";
import {ok, err, Result} from "neverthrow";
import * as targz from "targz"
import * as filesystem from "fs"
import * as path from "path"
import * as os from "os";

const COMPRESSION_EXTENSION = ".tgz"
const GRPC_DATA_TRANSFER_LIMIT = 3999000 //3.999 Mb. 1kb wiggle room. 1kb being about the size of a 2 paragraph readme.
const COMPRESSION_TEMP_FOLDER_PREFIX = "temp-node-archiver-compression-"
export class NodeTgzArchiver implements GenericTgzArchiver{

     public async createTgzByteArray(pathToArchive: string): Promise<Result<Uint8Array, Error>> {
         //Check if it exists
         if (!filesystem.existsSync(pathToArchive)) {
             return err(new Error("The file or folder you want to upload does not exist."))
         }

         //Make directory for usage.
         const osTempDirpath = os.tmpdir()
         const tempPathResponse = await filesystem.promises.mkdtemp(
             path.join(osTempDirpath, COMPRESSION_TEMP_FOLDER_PREFIX),
         ).then((folder: string) => {
             return ok(folder)
         }).catch((tempDirErr: Error) => {
             return err(tempDirErr)
         });
         if (tempPathResponse.isErr()) {
             return err(new Error("Failed to create temporary directory for file compression."))
         }

         const baseName = path.basename(pathToArchive) + COMPRESSION_EXTENSION
         const archiveOptions = {
             src: pathToArchive,
             dest: path.join(tempPathResponse.value, baseName),
         }

         const targzPromise: Promise<Result<null, Error>> = new Promise((resolve, unusedReject) => {
             targz.compress(archiveOptions, (callbackErr: string | Error | null) => {
                if (callbackErr !== null) {
                    if (typeof callbackErr === "string") {
                        resolve(err(new Error(callbackErr)))
                        return
                    // Duck-typing way to check if this is an Error type
                    } else if (callbackErr && callbackErr.stack && callbackErr.message) {
                        resolve(err(callbackErr))
                        return
                    } else {
                        resolve(err(new Error(
                            `Compression callback encountered an unknown error type '${callbackErr}'; ` +
                                `this should never happen.`)))
                        return
                    }
                }
                resolve(ok(null))
             });
         })
         const targzResult = await targzPromise
         if(targzResult.isErr()) {
             return err(targzResult.error)
         }

         if (!filesystem.existsSync(archiveOptions.dest)) {
             return err(new Error(`Your files were compressed but could not be found at '${archiveOptions.dest}'.`))
         }

         const stats = filesystem.statSync(archiveOptions.dest)
         if (stats.size >= GRPC_DATA_TRANSFER_LIMIT) {
             return err(new Error("The files you are trying to upload, which are now compressed, exceed or reach 4mb, " +
                 "a limit imposed by gRPC. Please reduce the total file size and ensure it can compress to a size below 4mb."))
         }

         if (stats.size <= 0) {
             return err(new Error("Something went wrong during compression. The compressed file size is 0 bytes."))
         }

         const data = filesystem.readFileSync(archiveOptions.dest)
         if(data.length != stats.size){
             return err(new Error(`Something went wrong while reading your recently compressed file '${baseName}'.` +
                 `The file size of ${stats.size} bytes and read size of ${data.length} bytes are not equal.`))
         }

         return ok(new Uint8Array(data.buffer))
    }
}