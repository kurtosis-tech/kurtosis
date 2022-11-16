//This is an object to represent a simple log line information
export class ServiceLog {
    //private readonly lineTime: Date; //TODO add the time from loki logs result
    private readonly content: string;

    constructor(content: string) {
        this.content = content;
    }

    public getContent():string {
        return this.content;
    }
}
