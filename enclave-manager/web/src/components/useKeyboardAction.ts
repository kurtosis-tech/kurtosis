import { useEffect } from "react";

export type KeyboardActions = "escape" | "find" | "omniFind" | "next";

export type OnCtrlPressHandlers = Partial<Record<KeyboardActions, () => void>>;

const eventIsType = (e: KeyboardEvent, type: KeyboardActions) => {
  const ctrlOrMeta = e.ctrlKey || e.metaKey;

  switch (type) {
    case "find":
      return ctrlOrMeta && e.keyCode === 70; // F
    case "next":
      return ctrlOrMeta && e.keyCode === 71; // G
    case "omniFind":
      return ctrlOrMeta && e.keyCode === 75; // K
    case "escape":
      return e.key === "Escape" || e.keyCode === 27;
  }
};

export const useKeyboardAction = (handlers: OnCtrlPressHandlers) => {
  useEffect(() => {
    const listener = function (e: KeyboardEvent) {
      for (const [handlerType, handler] of Object.entries(handlers)) {
        if (eventIsType(e, handlerType as KeyboardActions)) {
          e.preventDefault();
          handler();
          return;
        }
      }
    };
    window.addEventListener("keydown", listener);
    return () => window.removeEventListener("keydown", listener);
  }, [handlers]);
};
