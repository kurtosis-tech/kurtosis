import { isDefined, maybeParse } from "kurtosis-ui-components";
import { createContext, PropsWithChildren, useCallback, useContext, useEffect, useState } from "react";

type Settings = {
  ENABLE_EXPERIMENTAL_BUILD_ENCLAVE: boolean;
  SAVED_PACKAGES: string[];
};

export const settingKeys: { [k in keyof Settings]: k } = {
  ENABLE_EXPERIMENTAL_BUILD_ENCLAVE: "ENABLE_EXPERIMENTAL_BUILD_ENCLAVE",
  SAVED_PACKAGES: "SAVED_PACKAGES",
} as const;

const defaultSettings: Settings = { ENABLE_EXPERIMENTAL_BUILD_ENCLAVE: false, SAVED_PACKAGES: [] };

const SETTINGS_LOCAL_STORAGE_KEY = "kurtosis-settings";

export const storeSettings = (settings: Settings) => {
  localStorage.setItem(SETTINGS_LOCAL_STORAGE_KEY, JSON.stringify(settings));
};

export const loadSettings = (): Settings => {
  // TODO: Remove once confident all users have migrated from the old settings key
  const oldSavedPackages = localStorage.getItem("kurtosis-saved-packages");
  const migratedDefaultSettings = {
    ...defaultSettings,
    SAVED_PACKAGES: isDefined(oldSavedPackages)
      ? maybeParse(oldSavedPackages, defaultSettings.SAVED_PACKAGES)
      : defaultSettings.SAVED_PACKAGES,
  };

  const savedRawValue = localStorage.getItem(SETTINGS_LOCAL_STORAGE_KEY);

  if (!isDefined(savedRawValue)) {
    return migratedDefaultSettings;
  }

  return maybeParse(savedRawValue, migratedDefaultSettings);
};

type SettingsContextState = {
  settings: Settings;
  updateSetting: <S extends keyof Settings>(setting: S, value: Settings[S]) => void;
};

const SettingsContext = createContext<SettingsContextState>(null as any);

export const SettingsContextProvider = ({ children }: PropsWithChildren) => {
  const [settings, setSettings] = useState<Settings>(defaultSettings);

  const updateSetting = useCallback(<S extends keyof Settings>(setting: S, value: Settings[S]) => {
    setSettings((settings) => {
      const newSettings = { ...settings, [setting]: value };
      storeSettings(newSettings);
      return newSettings;
    });
  }, []);

  useEffect(() => {
    setSettings(loadSettings());
  }, []);

  return <SettingsContext.Provider value={{ settings, updateSetting }}>{children}</SettingsContext.Provider>;
};

export const useSettings = () => useContext(SettingsContext);
