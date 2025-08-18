package population

// GetHospitalTemplate returns a comprehensive hospital population template
func GetHospitalTemplate() *PopulationTemplate {
	return &PopulationTemplate{
		Domain:      "hospital",
		Description: "Regional hospital with beds, patients, staff, and medical records",
		BaseMetrics: map[string]MetricRatio{
			"patients": {
				Name:         "patients",
				Ratio:        5.0, // 5 patients per bed for quick testing
				Distribution: "poisson",
				MinValue:     10,
				MaxValue:     100,
				Description:  "Annual patient admissions per bed",
			},
			"providers": {
				Name:         "providers",
				Ratio:        0.8, // 0.8 providers per bed
				Distribution: "normal",
				MinValue:     10,
				MaxValue:     0,
				Description:  "Medical providers (doctors, nurses, specialists)",
			},
			"claims": {
				Name:         "claims",
				Ratio:        7.5, // 7.5 claims per bed for quick testing
				Distribution: "poisson",
				MinValue:     20,
				MaxValue:     200,
				Description:  "Healthcare claims (837 EDI)",
			},
			"prescriptions": {
				Name:         "prescriptions",
				Ratio:        12.0, // 12 prescriptions per bed for quick testing
				Distribution: "poisson",
				MinValue:     30,
				MaxValue:     300,
				Description:  "Pharmacy prescriptions (NCPDP)",
			},
			"procedures": {
				Name:         "procedures",
				Ratio:        2.5, // 2.5 procedures per bed for quick testing
				Distribution: "poisson",
				MinValue:     5,
				MaxValue:     50,
				Description:  "Medical procedures and surgeries",
			},
			"lab_results": {
				Name:         "lab_results",
				Ratio:        20.0, // 20 lab results per bed for quick testing
				Distribution: "poisson",
				MinValue:     50,
				MaxValue:     500,
				Description:  "Laboratory test results",
			},
		},
		Schemas: []SchemaRecommendation{
			{
				SchemaPath:   "test/schemas/medical/healthcare-claims-837.json",
				RecordType:   "claims",
				Priority:     "critical",
				Dependencies: []string{"patients", "providers"},
			},
			{
				SchemaPath:   "test/schemas/medical/rx-claims-ncpdp.json",
				RecordType:   "prescriptions",
				Priority:     "important",
				Dependencies: []string{"patients", "providers"},
			},
		},
		Relationships: []RelationshipRule{
			{
				ParentType:   "patients",
				ChildType:    "claims",
				Relationship: "one-to-many",
				Ratio:        1.5,
				Description:  "Each patient has multiple claims",
			},
			{
				ParentType:   "patients",
				ChildType:    "prescriptions",
				Relationship: "one-to-many",
				Ratio:        2.4,
				Description:  "Each patient has multiple prescriptions",
			},
			{
				ParentType:   "providers",
				ChildType:    "claims",
				Relationship: "one-to-many",
				Ratio:        93.75,
				Description:  "Each provider handles multiple claims",
			},
		},
	}
}

// GetBankTemplate returns a comprehensive bank population template
func GetBankTemplate() *PopulationTemplate {
	return &PopulationTemplate{
		Domain:      "bank",
		Description: "Community or regional bank with branches, customers, and transactions",
		BaseMetrics: map[string]MetricRatio{
			"customers": {
				Name:         "customers",
				Ratio:        4000.0, // 4000 customers per branch
				Distribution: "normal",
				MinValue:     1000,
				MaxValue:     0,
				Description:  "Bank customers per branch",
			},
			"accounts": {
				Name:         "accounts",
				Ratio:        6000.0, // 6000 accounts per branch (1.5 per customer)
				Distribution: "normal",
				MinValue:     1500,
				MaxValue:     0,
				Description:  "Bank accounts per branch",
			},
			"transactions": {
				Name:         "transactions",
				Ratio:        50000.0, // 50K transactions per branch monthly
				Distribution: "poisson",
				MinValue:     10000,
				MaxValue:     0,
				Description:  "Monthly transactions per branch",
			},
			"loans": {
				Name:         "loans",
				Ratio:        800.0, // 800 loans per branch
				Distribution: "normal",
				MinValue:     200,
				MaxValue:     0,
				Description:  "Active loans per branch",
			},
			"credit_cards": {
				Name:         "credit_cards",
				Ratio:        2400.0, // 2400 credit cards per branch
				Distribution: "normal",
				MinValue:     600,
				MaxValue:     0,
				Description:  "Active credit cards per branch",
			},
		},
		Schemas: []SchemaRecommendation{
			{
				SchemaPath:   "test/schemas/fintech/transactions.json",
				RecordType:   "transactions",
				Priority:     "critical",
				Dependencies: []string{"customers", "accounts"},
			},
		},
		Relationships: []RelationshipRule{
			{
				ParentType:   "customers",
				ChildType:    "accounts",
				Relationship: "one-to-many",
				Ratio:        1.5,
				Description:  "Each customer has multiple accounts",
			},
			{
				ParentType:   "accounts",
				ChildType:    "transactions",
				Relationship: "one-to-many",
				Ratio:        8.33,
				Description:  "Each account has multiple transactions",
			},
		},
	}
}

// GetRetailTemplate returns a comprehensive retail store population template
func GetRetailTemplate() *PopulationTemplate {
	return &PopulationTemplate{
		Domain:      "retail",
		Description: "Retail chain with stores, products, customers, and sales",
		BaseMetrics: map[string]MetricRatio{
			"products": {
				Name:         "products",
				Ratio:        5000.0, // 5000 products per store
				Distribution: "normal",
				MinValue:     1000,
				MaxValue:     0,
				Description:  "Product catalog per store",
			},
			"customers": {
				Name:         "customers",
				Ratio:        2000.0, // 2000 regular customers per store
				Distribution: "normal",
				MinValue:     500,
				MaxValue:     0,
				Description:  "Regular customers per store",
			},
			"orders": {
				Name:         "orders",
				Ratio:        10000.0, // 10K orders per store monthly
				Distribution: "poisson",
				MinValue:     2000,
				MaxValue:     0,
				Description:  "Monthly orders per store",
			},
			"inventory": {
				Name:         "inventory",
				Ratio:        5000.0, // 1:1 with products
				Distribution: "normal",
				MinValue:     1000,
				MaxValue:     0,
				Description:  "Inventory records per store",
			},
			"employees": {
				Name:         "employees",
				Ratio:        25.0, // 25 employees per store
				Distribution: "normal",
				MinValue:     5,
				MaxValue:     100,
				Description:  "Store employees",
			},
		},
		Schemas: []SchemaRecommendation{
			{
				SchemaPath:   "test/schemas/ecommerce/products.json",
				RecordType:   "products",
				Priority:     "critical",
				Dependencies: []string{},
			},
			{
				SchemaPath:   "test/schemas/x12/purchase-order-850.json",
				RecordType:   "orders",
				Priority:     "important",
				Dependencies: []string{"products", "customers"},
			},
		},
		Relationships: []RelationshipRule{
			{
				ParentType:   "customers",
				ChildType:    "orders",
				Relationship: "one-to-many",
				Ratio:        5.0,
				Description:  "Each customer places multiple orders",
			},
			{
				ParentType:   "products",
				ChildType:    "inventory",
				Relationship: "one-to-one",
				Ratio:        1.0,
				Description:  "Each product has inventory record",
			},
		},
	}
}

// GetEcommerceTemplate returns a comprehensive e-commerce platform template
func GetEcommerceTemplate() *PopulationTemplate {
	return &PopulationTemplate{
		Domain:      "ecommerce",
		Description: "E-commerce platform with users, products, orders, and transactions",
		BaseMetrics: map[string]MetricRatio{
			"products": {
				Name:         "products",
				Ratio:        0.1, // 0.1 products per user (10K users = 1K products)
				Distribution: "normal",
				MinValue:     500,
				MaxValue:     0,
				Description:  "Product catalog size relative to user base",
			},
			"orders": {
				Name:         "orders",
				Ratio:        2.5, // 2.5 orders per user annually
				Distribution: "poisson",
				MinValue:     1000,
				MaxValue:     0,
				Description:  "Annual orders per user",
			},
			"reviews": {
				Name:         "reviews",
				Ratio:        0.8, // 0.8 reviews per user annually
				Distribution: "poisson",
				MinValue:     100,
				MaxValue:     0,
				Description:  "Product reviews per user",
			},
			"cart_sessions": {
				Name:         "cart_sessions",
				Ratio:        12.0, // 12 cart sessions per user annually
				Distribution: "poisson",
				MinValue:     1000,
				MaxValue:     0,
				Description:  "Shopping cart sessions per user",
			},
			"payments": {
				Name:         "payments",
				Ratio:        2.5, // 1:1 with orders
				Distribution: "poisson",
				MinValue:     1000,
				MaxValue:     0,
				Description:  "Payment transactions per user",
			},
		},
		Schemas: []SchemaRecommendation{
			{
				SchemaPath:   "test/schemas/ecommerce/products.json",
				RecordType:   "products",
				Priority:     "critical",
				Dependencies: []string{},
			},
		},
		Relationships: []RelationshipRule{
			{
				ParentType:   "users",
				ChildType:    "orders",
				Relationship: "one-to-many",
				Ratio:        2.5,
				Description:  "Each user places multiple orders",
			},
			{
				ParentType:   "orders",
				ChildType:    "payments",
				Relationship: "one-to-one",
				Ratio:        1.0,
				Description:  "Each order has a payment",
			},
		},
	}
}

// GetInsuranceTemplate returns a comprehensive insurance company template
func GetInsuranceTemplate() *PopulationTemplate {
	return &PopulationTemplate{
		Domain:      "insurance",
		Description: "Insurance company with policies, claims, and members",
		BaseMetrics: map[string]MetricRatio{
			"members": {
				Name:         "members",
				Ratio:        1.0, // 1 member per policyholder
				Distribution: "normal",
				MinValue:     1,
				MaxValue:     0,
				Description:  "Insurance members per policyholder",
			},
			"policies": {
				Name:         "policies",
				Ratio:        1.2, // 1.2 policies per policyholder
				Distribution: "normal",
				MinValue:     1,
				MaxValue:     0,
				Description:  "Active policies per policyholder",
			},
			"claims": {
				Name:         "claims",
				Ratio:        0.5, // 0.5 claims per policyholder annually
				Distribution: "poisson",
				MinValue:     1,
				MaxValue:     0,
				Description:  "Annual claims per policyholder",
			},
			"agents": {
				Name:         "agents",
				Ratio:        0.1, // 0.1 agents per policyholder (1 agent per 10 policyholders)
				Distribution: "normal",
				MinValue:     1,
				MaxValue:     0,
				Description:  "Insurance agents per policyholder",
			},
			"payments": {
				Name:         "payments",
				Ratio:        3.0, // 3 payments per policyholder annually
				Distribution: "normal",
				MinValue:     1,
				MaxValue:     0,
				Description:  "Annual payments per policyholder",
			},
		},
		Schemas: []SchemaRecommendation{
			{
				SchemaPath:   "test/schemas/insurance/claims.json",
				RecordType:   "claims",
				Priority:     "critical",
				Dependencies: []string{"members", "agents"},
			},
		},
		Relationships: []RelationshipRule{
			{
				ParentType:   "members",
				ChildType:    "policies",
				Relationship: "one-to-many",
				Ratio:        1.2,
				Description:  "Each member has multiple policies",
			},
			{
				ParentType:   "members",
				ChildType:    "claims",
				Relationship: "one-to-many",
				Ratio:        0.3,
				Description:  "Each member has occasional claims",
			},
		},
	}
}
