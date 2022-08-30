/*
 * Copyright (c) 2022 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

import "neverthrow"
import {GenericTgzArchiver} from "./generic_tgz_archiver";
import {ok, err, Result} from "neverthrow";
import * as filesystem from "fs"
import * as path from "path"
import * as os from "os";
import * as tar from "tar";

const COMPRESSION_EXTENSION = ".tgz"
const GRPC_DATA_TRANSFER_LIMIT = 3999000 //3.999 Mb. 1kb wiggle room. 1kb being about the size of a 2 paragraph readme.
const COMPRESSION_TEMP_FOLDER_PREFIX = "temp-node-archiver-compression-"
export class NodeTgzArchiver implements GenericTgzArchiver{

     public async createTgzByteArray(pathToArchive: string): Promise<Result<Uint8Array, Error>> {
         //Check if it exists
         if (!filesystem.existsSync(pathToArchive)) {
             return err(new Error(`Path '${pathToArchive}' does not exist.`))
         }
         if (pathToArchive === "/") {
             return err(new Error("Cannot archive the root directory"))
         }

         const srcParentDirpath = path.dirname(pathToArchive)
         const srcFilename = path.basename(pathToArchive)
         const isPathToArchiveDirectory = filesystem.lstatSync(pathToArchive).isDirectory()

         //Make directory for usage.
         const osTempDirpath = os.tmpdir()
         const tempDirpathPrefix = path.join(osTempDirpath, COMPRESSION_TEMP_FOLDER_PREFIX)
         const tempDirpathResult = await filesystem.promises.mkdtemp(
             tempDirpathPrefix,
         ).then((folder: string) => {
             return ok(folder)
         }).catch((tempDirErr: Error) => {
             return err(tempDirErr)
         });
         if (tempDirpathResult.isErr()) {
             return err(tempDirpathResult.error)
         }
         const tempDirpath = tempDirpathResult.value
         const destFilename = srcFilename + COMPRESSION_EXTENSION
         const destFilepath = path.join(tempDirpath, destFilename)

         const filenamesToUpload = isPathToArchiveDirectory ? filesystem.readdirSync(pathToArchive) : [srcFilename]
         if (filenamesToUpload.length == 0) {
            return err(new Error(`The directory '${pathToArchive}' you are trying to upload is empty`))
         }
         const targzPromise = tar.create(
             {
                 cwd: isPathToArchiveDirectory? pathToArchive : srcParentDirpath,
                 gzip: true,
                 file: destFilepath,
             },
             filenamesToUpload,
         ).then((_) => {
             return ok(null)
         }).catch((err: any) => {
             // Use duck-typing to detect Error types
             if (err && err.stack && err.message) {
                 return err(err as Error);
             }
             return err(new Error(`A non-Error object '${err}' was thrown when compressing '${pathToArchive}' to '${destFilepath}'`))
         })
         const targzResult = await targzPromise
         if(targzResult.isErr()) {
             return err(targzResult.error)
         }

         if (!filesystem.existsSync(destFilepath)) {
             return err(new Error(`Your files were compressed but could not be found at '${destFilepath}'.`))
         }

         const stats = filesystem.statSync(destFilepath)
         if (stats.size >= GRPC_DATA_TRANSFER_LIMIT) {
             return err(new Error("The files you are trying to upload, which are now compressed, exceed or reach 4mb, " +
                 "a limit imposed by gRPC. Please reduce the total file size and ensure it can compress to a size below 4mb."))
         }

         if (stats.size <= 0) {
             return err(new Error("Something went wrong during compression. The compressed file size is 0 bytes."))
         }

         const data = filesystem.readFileSync(destFilepath)
         if(data.length != stats.size){
             return err(new Error(`Something went wrong while reading your recently compressed file '${destFilename}'.` +
                 `The file size of ${stats.size} bytes and read size of ${data.length} bytes are not equal.`))
         }

         return ok(data)
    }
}