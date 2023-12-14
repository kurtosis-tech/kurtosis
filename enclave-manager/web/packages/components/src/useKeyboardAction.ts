import { useEffect } from "react";
import { isDefined } from "./utils";

export type KeyboardActions = "escape" | "find" | "omniFind" | "enter" | "shift-enter";

export type OnCtrlPressHandlers = Partial<Record<KeyboardActions, () => void>>;

const getEventType = (e: KeyboardEvent): KeyboardActions | null => {
  const ctrlOrMeta = e.ctrlKey || e.metaKey;

  if (ctrlOrMeta && e.keyCode === 70) {
    // F
    return "find";
  }
  if (e.shiftKey && e.keyCode === 13) {
    // shift + enter
    return "shift-enter";
  }
  if (e.keyCode === 13) {
    // enter
    return "enter";
  }
  if (ctrlOrMeta && e.keyCode === 75) {
    // K
    return "omniFind";
  }
  if (e.key === "Escape" || e.keyCode === 27) {
    return "escape";
  }
  return null;
};

export const useKeyboardAction = (handlers: OnCtrlPressHandlers) => {
  useEffect(() => {
    const listener = function (e: KeyboardEvent) {
      const eventType = getEventType(e);
      const handler = isDefined(eventType) ? handlers[eventType] : null;
      if (isDefined(handler)) {
        e.preventDefault();
        handler();
        return;
      }
    };
    window.addEventListener("keydown", listener);
    return () => window.removeEventListener("keydown", listener);
  }, [handlers]);
};
