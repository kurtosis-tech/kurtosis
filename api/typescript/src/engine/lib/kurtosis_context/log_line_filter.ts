import {LogLineOperator} from "./log_line_operator";

const DEFAULT_OPERATOR: LogLineOperator = LogLineOperator.DoesContainText;

export class LogLineFilter {

    private operator: LogLineOperator = DEFAULT_OPERATOR;
    private textPattern: string = "";

    public getOperator(): LogLineOperator {
        return this.operator;
    }

    public getTextPattern(): string {
        return this.textPattern;
    }

    public static NewDoesContainTextLogLineFilter(text: string): LogLineFilter {
        const operator: LogLineOperator = LogLineOperator.DoesContainText;
        return this.newLogLineFilter(operator, text);
    }

    public static NewDoesNotContainTextLogLineFilter(text: string): LogLineFilter {
        const operator: LogLineOperator = LogLineOperator.DoesNotContainText;
        return this.newLogLineFilter(operator, text);
    }

    public static NewDoesContainMatchRegexLogLineFilter(regex: string): LogLineFilter {
        const operator: LogLineOperator = LogLineOperator.DoesContainMatchRegex;
        return this.newLogLineFilter(operator, regex);
    }

    public static NewDoesNotContainMatchRegexLogLineFilter(regex: string): LogLineFilter {
        const operator: LogLineOperator = LogLineOperator.DoesNotContainMatchRegex;
        return this.newLogLineFilter(operator, regex);
    }

    private static newLogLineFilter(operator: LogLineOperator, textPattern: string): LogLineFilter {
        const filter: LogLineFilter = new LogLineFilter();
        filter.operator = operator;
        filter.textPattern = textPattern;
        return filter;
    }
}
