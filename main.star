IMAGE_ARG_KEY = "image"
SERVICE_NAME_ARG_KEY = "name"
DATABASE_ARG_KEY = "database"
USER_ARG_KEY = "user"
PASSWORD_ARG_KEY = "password"

# comma separated postgres config key=value pairs
POSTGRES_CONFIG_KEY = "postgres_config"

SEED_FILE_ARTIFACT_ARG_KEY = "seed_file_artifact"
CONFIG_FILE_ARTIFACT_ARG_KEY = "config_file_artifact"

PORT_NAME = "postgresql"
APPLICATION_PROTOCOL = "postgresql"

CONFIG_FILE_MOUNT_DIRPATH = "/config"
SEED_FILE_MOUNT_PATH = "/docker-entrypoint-initdb.d"

CONFIG_FILENAME = "postgresql.conf"  # Expected to be in the artifact

def run(plan, args={}):
    image = args.get(IMAGE_ARG_KEY, "postgres:alpine")
    service_name = args.get(SERVICE_NAME_ARG_KEY, "postgres")
    user = args.get(USER_ARG_KEY, "postgres")
    password = args.get(PASSWORD_ARG_KEY, "MyPassword1!")
    database = args.get(DATABASE_ARG_KEY, "postgres")
    postgres_configs = args.get(POSTGRES_CONFIG_KEY, [])
    config_file_artifact_name = args.get(CONFIG_FILE_ARTIFACT_ARG_KEY, "")
    seed_file_artifact_name = args.get(SEED_FILE_ARTIFACT_ARG_KEY, "")

    cmd = []
    files = {}
    if config_file_artifact_name != "":
        config_filepath = CONFIG_FILE_MOUNT_DIRPATH + "/" + CONFIG_FILENAME
        cmd += ["-c", "config_file=" + config_filepath]
        files[CONFIG_FILE_MOUNT_DIRPATH] = config_file_artifact_name

    # append cmd with postgres config overrides passed by users
    if len(postgres_configs) > 0:
        for postgres_config in postgres_configs:
            cmd += ["-c", postgres_config]

    if seed_file_artifact_name != "":
        files[SEED_FILE_MOUNT_PATH] = seed_file_artifact_name

    service = plan.add_service(
                      name = "busy",
                      config = ServiceConfig(
                          image = image,
                          ports = {
                              PORT_NAME: PortSpec(
                                  number = 5432,
                                  application_protocol = APPLICATION_PROTOCOL,
                              )
                          },
                          cmd = cmd,
                          env_vars = {
                              "POSTGRES_DB": database,
                              "POSTGRES_USER": user,
                              "POSTGRES_PASSWORD": password,
                          },
                          files = files,
                      )
                  )

    plan.exec(
        service_name="busy",
		recipe = ExecRecipe(
			command = ["wget", "-O", "./output.txt", "https://raw.githubusercontent.com/ethereum/c-kzg-4844/main/src/trusted_setup.txt"]
		)
    )
