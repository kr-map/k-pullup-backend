package csw.chulbongkr.config.custom;

import lombok.Getter;
import lombok.Setter;
import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.context.annotation.Configuration;

import java.time.Duration;

@Getter
@Setter
@Configuration
@ConfigurationProperties(prefix = "chulbong")
public class AppConfig {
    private Duration tokenExpirationInterval;
    private String clientAddress;
    private String clientRedirectEndpoint;
    private String tokenCookie;
    private Smtp smtp;
    private String frontendResetRouter;
    private String encryptionKey;
    private String ckUrl;
    private String naverEmailVerifyUrl;
    private String guktoApiKey;

    @Getter
    @Setter
    public static class Smtp {
        private String server;
        private int port;
        private String username;
        private String password;
    }
}