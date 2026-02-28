from pydantic import Field
from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    embed_model_name: str = Field(
        "all-MiniLM-L6-v2", env="EMBED_MODEL_NAME")
    rerank_model_name: str = Field(
        "cross-encoder/ms-marco-MiniLM-L-6-v2", env="RERANK_MODEL_NAME")
    reader_model_name: str = Field(
        "distilbert-base-cased-distilled-squad", env="READER_MODEL_NAME")
    max_text_length: int = Field(5000, env="MAX_TEXT_LENGTH")

    port: int = Field(5001, env="PORT")

    log_level: str = Field("INFO", env="LOG_LEVEL")


    model_config = SettingsConfigDict(
        env_file=".env",
        extra="ignore"  # Add fields from .env
    )


settings = Settings()

