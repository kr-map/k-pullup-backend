package csw.chulbongkr;

import csw.chulbongkr.service.search.LuceneIndexerService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.CommandLineRunner;
import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.context.annotation.Bean;
import org.springframework.scheduling.annotation.EnableAsync;

@EnableAsync
@SpringBootApplication
public class ChulbongKrApplication {

	@Autowired
	private LuceneIndexerService luceneIndexerService;

	public static void main(String[] args) {
		SpringApplication application = new SpringApplication(ChulbongKrApplication.class);
		application.setLazyInitialization(true);
		application.run(args);
	}

	@Bean
	public CommandLineRunner commandLineRunner() {
		return args -> {
			// Ensure the LuceneIndexerService bean is initialized and indexing data
			luceneIndexerService.ensureInitialized();
		};
	}
}

