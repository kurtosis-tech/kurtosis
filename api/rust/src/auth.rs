use std::env;
use std::fs;
use std::path::PathBuf;

use tonic::metadata::MetadataValue;
use tonic::service::Interceptor;
use tonic::{Request, Status};

pub const AUTH_TOKEN_ENV_VAR_NAME: &str = "KURTOSIS_API_AUTH_TOKEN";

const APPLICATION_DIRNAME: &str = "kurtosis";
const TOKEN_FILE_NAME: &str = "api-auth-token";
const GRPC_METADATA_KEY: &str = "authorization";
const BEARER_PREFIX: &str = "Bearer ";

fn config_home_dirpath() -> Option<PathBuf> {
    if let Some(xdg_config_home) = env::var_os("XDG_CONFIG_HOME").filter(|value| !value.is_empty()) {
        return Some(PathBuf::from(xdg_config_home));
    }
    if cfg!(target_os = "windows") {
        return env::var_os("LOCALAPPDATA").map(PathBuf::from);
    }
    let home_dirpath = PathBuf::from(env::var_os("HOME")?);
    if cfg!(target_os = "macos") {
        return Some(home_dirpath.join("Library").join("Application Support"));
    }
    Some(home_dirpath.join(".config"))
}

pub fn get_auth_token() -> Option<String> {
    if let Ok(token) = env::var(AUTH_TOKEN_ENV_VAR_NAME) {
        let token = token.trim();
        if !token.is_empty() {
            return Some(token.to_string());
        }
    }
    let token_filepath = config_home_dirpath()?.join(APPLICATION_DIRNAME).join(TOKEN_FILE_NAME);
    let token = fs::read_to_string(token_filepath).ok()?;
    let token = token.trim();
    if token.is_empty() {
        return None;
    }
    Some(token.to_string())
}

#[derive(Clone, Debug)]
pub struct AuthInterceptor {
    token: Option<String>,
}

impl AuthInterceptor {
    pub fn new(token: Option<String>) -> Self {
        AuthInterceptor { token }
    }

    pub fn from_local_config() -> Self {
        AuthInterceptor::new(get_auth_token())
    }
}

impl Interceptor for AuthInterceptor {
    fn call(&mut self, mut request: Request<()>) -> Result<Request<()>, Status> {
        if let Some(token) = &self.token {
            let header_value = MetadataValue::try_from(format!("{}{}", BEARER_PREFIX, token))
                .map_err(|_| Status::unauthenticated("the Kurtosis API auth token is not a valid header value"))?;
            request.metadata_mut().insert(GRPC_METADATA_KEY, header_value);
        }
        Ok(request)
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn interceptor_adds_bearer_token() {
        let mut interceptor = AuthInterceptor::new(Some("test-token".to_string()));
        let request = interceptor.call(Request::new(())).unwrap();
        assert_eq!(request.metadata().get(GRPC_METADATA_KEY).unwrap(), "Bearer test-token");
    }

    #[test]
    fn interceptor_without_token_adds_nothing() {
        let mut interceptor = AuthInterceptor::new(None);
        let request = interceptor.call(Request::new(())).unwrap();
        assert!(request.metadata().get(GRPC_METADATA_KEY).is_none());
    }
}
