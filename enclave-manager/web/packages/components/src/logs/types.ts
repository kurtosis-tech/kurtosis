import { DateTime } from "luxon";

export type LogStatus = "info" | "error";

export type LogLineMessage = {
  status?: LogStatus;
  message?: string;
  timestamp?: DateTime;
};
