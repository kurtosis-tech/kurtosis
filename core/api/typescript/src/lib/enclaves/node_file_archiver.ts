/*
 * Copyright (c) 2022 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */
import "targz";
import "fs";
import "path";
import "neverthrow"
import {GenericTgzArchiver} from "./generic_tgz_archiver";
import {ok, err, Result} from "neverthrow";

const COMPRESSION_EXTENSION = ".tgz"
const GRPC_DATA_TRANSFER_LIMIT = 3999000 //3.999 Mb. 1kb wiggle room. 1kb being about the size of a 2 paragraph readme.
const COMPRESSION_TEMP_FOLDER_PREFIX = "temp-node-archiver-compression-"
export class NodeFileArchiver implements GenericTgzArchiver{

     public async createTgz(pathToArchive: string): Promise<Result<Uint8Array, Error>> {
         const targz = require("targz")
         const filesystemPromises = require("fs").promises
         const filesystem = require("fs")
         const path = require("path")

         //Check if it exists
         if(!filesystem.existsSync(pathToArchive)) {
             return err(new Error("The file or folder you want to upload does not exist."))
         }

         //Make directory for usage.
         var absoluteTarPath
         filesystemPromises.mkdtemp(COMPRESSION_TEMP_FOLDER_PREFIX)
             .then((folder : string) =>{
                 absoluteTarPath = folder
             })
             .catch((tempDirErr: Error)=>{
                 return err(tempDirErr)
             });

         const baseName = path.basename(pathToArchive) + COMPRESSION_EXTENSION
         const archiveOptions  = {
             src: pathToArchive,
             dest: path.join(absoluteTarPath,baseName),
         }

         var error : Error | string | null = null
         targz.compress(archiveOptions, (compressErr: Error) => { error = compressErr })
         if(error != null){
             return err(error)
         }

         if (!filesystem.existsSync(archiveOptions.dest)){
             return err(new Error(`Your files were compressed but could not be found at '${archiveOptions.dest}'.`))
         }

         const stats = filesystem.statSync(archiveOptions.dest)
         if(stats.size >= GRPC_DATA_TRANSFER_LIMIT){
             return err(new Error("The files you are trying to upload, which are now compressed, exceed or reach 4mb, " +
                 "a limit imposed by gRPC. Please reduce the total file size and ensure it can compress to a size below 4mb."))
         }
         const data = filesystem.readFileSync(archiveOptions.dest)
         if(data.length != stats.size){
             return err(new Error(`Something went wrong while reading your recently compressed file ${baseName}.` +
             `The file size of ${stats.size} bytes and read size of ${data.length} bytes are not equal.`))
         }

         return ok(data.buffer)
    }
}