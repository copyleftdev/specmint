# Ollama Setup and Validation

Set up and validate Ollama integration for SpecMint with qwen2.5:latest model.

## Steps

1. **Install Ollama (if not present)**
   ```bash
   # Check if Ollama is installed
   which ollama || {
     echo "Installing Ollama..."
     curl -fsSL https://ollama.ai/install.sh | sh
   }
   ```

2. **Start Ollama Service**
   ```bash
   # Start Ollama server
   ollama serve &
   
   # Wait for service to be ready
   sleep 5
   
   # Verify service is running
   curl -s http://localhost:11434/api/tags || {
     echo "Ollama service not responding"
     exit 1
   }
   ```

3. **Pull Required Model**
   ```bash
   # Pull qwen2.5:latest model (7.6B, Q4_K_M quantization)
   ollama pull qwen2.5:latest
   
   # Verify model is available
   ollama list | grep qwen2.5 || {
     echo "Failed to pull qwen2.5:latest"
     exit 1
   }
   ```

4. **Test Model Generation**
   ```bash
   # Test basic generation
   curl -X POST http://localhost:11434/api/generate \
     -H "Content-Type: application/json" \
     -d '{
       "model": "qwen2.5:latest",
       "prompt": "Generate a JSON object with name and age fields",
       "stream": false,
       "options": {
         "seed": 12345,
         "temperature": 0.1
       }
     }'
   ```

5. **Validate SpecMint Integration**
   ```bash
   # Test SpecMint can connect to Ollama
   specmint doctor --ollama-only
   
   # Test LLM field generation
   specmint generate \
     --schema ./test/simple_schema.json \
     --count 5 \
     --seed 12345 \
     --llm-mode fields \
     --out ./test-ollama-integration
   ```

6. **Performance Validation**
   ```bash
   # Test concurrent requests
   specmint benchmark \
     --schema ./test/simple_schema.json \
     --count 100 \
     --llm-workers 3 \
     --llm.max_rps 3 \
     --seed 12345
   ```

7. **Create Test Schema (if needed)**
   ```bash
   cat > ./test/simple_schema.json << 'SCHEMA'
   {
     "$schema": "https://json-schema.org/draft/2020-12/schema",
     "type": "object",
     "properties": {
       "id": {"type": "integer"},
       "name": {
         "type": "string",
         "description": "llm: Generate a realistic person name"
       },
       "email": {"type": "string", "format": "email"},
       "bio": {
         "type": "string",
         "x-llm": true,
         "description": "A short professional biography"
       }
     },
     "required": ["id", "name", "email"]
   }
   SCHEMA
   ```

## Success Criteria
- Ollama service responds to health checks
- qwen2.5:latest model is available and functional
- SpecMint can successfully connect and generate LLM-enhanced data
- Performance targets are met (p95 < 2s for LLM calls)
- Generated data passes schema validation
