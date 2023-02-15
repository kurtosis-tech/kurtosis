/*
 * Copyright (c) 2022 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */


//This interface was created intentionally in order to separate the imports("path" or "path-browserify") 
//depending where the App is executed. 
//Otherwise importing "path" in Web environment makes webpack's compiler throw an error.
export interface GenericPathJoiner {
    join(firstPath: string, secondPath: string): string
}