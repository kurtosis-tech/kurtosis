/*
 * Copyright (c) 2022 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */
import "targz";
import "fs";
import "os";
import "path";
import "neverthrow"
import {GenericTgzArchiver} from "./generic_tgz_archiver";
import {ok, err, Result, Err} from "neverthrow";

const COMPRESSION_EXTENSION = ".tgz"

export class NodeFileArchiver implements GenericTgzArchiver{

     public async createTgz(pathToArchive: string): Promise<Result<Uint8Array, Error>> {
         const targz = require("targz")
         const filesystem = require("fs")
         const os = require("os")
         const path = require("path")

         //Check if it exists
         if(!filesystem.existsSync(pathToArchive)) {
             return err(new Error("The file or folder you want to upload does not exist."))
         }

         //Make directory for usage.
         var absoluteTarPath : string = ""
         var tempDirectoryErr : Error | null = null
         //TODO FS.promises.make
         filesystem.mkdtemp(os.tmpdir(),  (tempDirError : Error, folder: string) => {
             tempDirectoryErr = tempDirError
             absoluteTarPath = path.join(folder, path.basename(pathToArchive)) ;
         });

         if (tempDirectoryErr != null){
             return err(tempDirectoryErr)
         }

         const baseName = path.basename(pathToArchive) + COMPRESSION_EXTENSION
         const options  = {
             src: pathToArchive,
             dest: path.join(absoluteTarPath,baseName),
         }

         var error : Error | string | null = null
         targz.compress(options, function(compressErr: Error) { error = compressErr })
         if(error != null){
             return err(error)
         }

         if (!filesystem.existsSync(options.dest)){
             return err(new Error(`Your files were compressed but could not be found at ${options.dest}.`))
         }

         //Convert to bytes.
         //Return

         return err(new Error("Node File Archiver has not been implemented yet."))
    }
}