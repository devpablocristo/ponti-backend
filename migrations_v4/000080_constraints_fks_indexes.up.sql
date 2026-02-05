-- ========================================
-- MIGRATION 000080 CONSTRAINTS FKS INDEXES (UP)
-- ========================================
-- Nota: SQL en inglés, comentarios en español

BEGIN;

-- PRIMARY KEYS
ALTER TABLE ONLY public.users
    ADD CONSTRAINT pk_users PRIMARY KEY (id);
ALTER TABLE ONLY public.business_parameters
    ADD CONSTRAINT pk_business_parameters PRIMARY KEY (id);
ALTER TABLE ONLY public.fx_rates
    ADD CONSTRAINT pk_fx_rates PRIMARY KEY (id);
ALTER TABLE ONLY public.customers
    ADD CONSTRAINT pk_customers PRIMARY KEY (id);
ALTER TABLE ONLY public.campaigns
    ADD CONSTRAINT pk_campaigns PRIMARY KEY (id);
ALTER TABLE ONLY public.projects
    ADD CONSTRAINT pk_projects PRIMARY KEY (id);
ALTER TABLE ONLY public.managers
    ADD CONSTRAINT pk_managers PRIMARY KEY (id);
ALTER TABLE ONLY public.project_managers
    ADD CONSTRAINT pk_project_managers PRIMARY KEY (project_id, manager_id);
ALTER TABLE ONLY public.lease_types
    ADD CONSTRAINT pk_lease_types PRIMARY KEY (id);
ALTER TABLE ONLY public.fields
    ADD CONSTRAINT pk_fields PRIMARY KEY (id);
ALTER TABLE ONLY public.lots
    ADD CONSTRAINT pk_lots PRIMARY KEY (id);
ALTER TABLE ONLY public.lot_dates
    ADD CONSTRAINT pk_lot_dates PRIMARY KEY (id);
ALTER TABLE ONLY public.crops
    ADD CONSTRAINT pk_crops PRIMARY KEY (id);
ALTER TABLE ONLY public.labor_types
    ADD CONSTRAINT pk_labor_types PRIMARY KEY (id);
ALTER TABLE ONLY public.labor_categories
    ADD CONSTRAINT pk_labor_categories PRIMARY KEY (id);
ALTER TABLE ONLY public.labors
    ADD CONSTRAINT pk_labors PRIMARY KEY (id);
ALTER TABLE ONLY public.workorders
    ADD CONSTRAINT pk_workorders PRIMARY KEY (id);
ALTER TABLE ONLY public.workorder_items
    ADD CONSTRAINT pk_workorder_items PRIMARY KEY (id);
ALTER TABLE ONLY public.types
    ADD CONSTRAINT pk_types PRIMARY KEY (id);
ALTER TABLE ONLY public.categories
    ADD CONSTRAINT pk_categories PRIMARY KEY (id);
ALTER TABLE ONLY public.supplies
    ADD CONSTRAINT pk_supplies PRIMARY KEY (id);
ALTER TABLE ONLY public.stocks
    ADD CONSTRAINT pk_stocks PRIMARY KEY (id);
ALTER TABLE ONLY public.supply_movements
    ADD CONSTRAINT pk_supply_movements PRIMARY KEY (id);
ALTER TABLE ONLY public.providers
    ADD CONSTRAINT pk_providers PRIMARY KEY (id);
ALTER TABLE ONLY public.investors
    ADD CONSTRAINT pk_investors PRIMARY KEY (id);
ALTER TABLE ONLY public.project_investors
    ADD CONSTRAINT pk_project_investors PRIMARY KEY (project_id, investor_id);
ALTER TABLE ONLY public.crop_commercializations
    ADD CONSTRAINT pk_crop_commercializations PRIMARY KEY (id);
ALTER TABLE ONLY public.admin_cost_investors
    ADD CONSTRAINT pk_admin_cost_investors PRIMARY KEY (project_id, investor_id);
ALTER TABLE ONLY public.field_investors
    ADD CONSTRAINT pk_field_investors PRIMARY KEY (field_id, investor_id);
ALTER TABLE ONLY public.project_dollar_values
    ADD CONSTRAINT pk_project_dollar_values PRIMARY KEY (id);
ALTER TABLE ONLY public.invoices
    ADD CONSTRAINT pk_invoices PRIMARY KEY (id);

-- UNIQUE CONSTRAINTS
ALTER TABLE ONLY public.users
    ADD CONSTRAINT uq_users_username UNIQUE (username);
ALTER TABLE ONLY public.business_parameters
    ADD CONSTRAINT uq_business_parameters_key UNIQUE (key);
ALTER TABLE ONLY public.fx_rates
    ADD CONSTRAINT uq_fx_rates_pair_date UNIQUE (currency_pair, effective_date);
ALTER TABLE ONLY public.customers
    ADD CONSTRAINT uq_customers_name UNIQUE (name);
ALTER TABLE ONLY public.campaigns
    ADD CONSTRAINT uq_campaigns_name UNIQUE (name);
ALTER TABLE ONLY public.managers
    ADD CONSTRAINT uq_managers_name UNIQUE (name);
ALTER TABLE ONLY public.lease_types
    ADD CONSTRAINT uq_lease_types_name UNIQUE (name);
ALTER TABLE ONLY public.crops
    ADD CONSTRAINT uq_crops_name UNIQUE (name);
ALTER TABLE ONLY public.labor_types
    ADD CONSTRAINT uq_labor_types_name UNIQUE (name);
ALTER TABLE ONLY public.types
    ADD CONSTRAINT uq_types_name UNIQUE (name);
ALTER TABLE ONLY public.providers
    ADD CONSTRAINT uq_providers_name UNIQUE (name);
ALTER TABLE ONLY public.investors
    ADD CONSTRAINT uq_investors_name UNIQUE (name);
ALTER TABLE ONLY public.project_dollar_values
    ADD CONSTRAINT uq_project_dollar_values_period UNIQUE (project_id, year, month);
ALTER TABLE ONLY public.invoices
    ADD CONSTRAINT uq_invoices_work_order UNIQUE (work_order_id);

-- FOREIGN KEYS (auditoría)
ALTER TABLE ONLY public.customers
    ADD CONSTRAINT fk_customers_created_by FOREIGN KEY (created_by) REFERENCES public.users(id);
ALTER TABLE ONLY public.customers
    ADD CONSTRAINT fk_customers_updated_by FOREIGN KEY (updated_by) REFERENCES public.users(id);
ALTER TABLE ONLY public.customers
    ADD CONSTRAINT fk_customers_deleted_by FOREIGN KEY (deleted_by) REFERENCES public.users(id);

ALTER TABLE ONLY public.campaigns
    ADD CONSTRAINT fk_campaigns_created_by FOREIGN KEY (created_by) REFERENCES public.users(id);
ALTER TABLE ONLY public.campaigns
    ADD CONSTRAINT fk_campaigns_updated_by FOREIGN KEY (updated_by) REFERENCES public.users(id);
ALTER TABLE ONLY public.campaigns
    ADD CONSTRAINT fk_campaigns_deleted_by FOREIGN KEY (deleted_by) REFERENCES public.users(id);

ALTER TABLE ONLY public.projects
    ADD CONSTRAINT fk_projects_customer FOREIGN KEY (customer_id) REFERENCES public.customers(id);
ALTER TABLE ONLY public.projects
    ADD CONSTRAINT fk_projects_campaign FOREIGN KEY (campaign_id) REFERENCES public.campaigns(id);
ALTER TABLE ONLY public.projects
    ADD CONSTRAINT fk_projects_created_by FOREIGN KEY (created_by) REFERENCES public.users(id);
ALTER TABLE ONLY public.projects
    ADD CONSTRAINT fk_projects_updated_by FOREIGN KEY (updated_by) REFERENCES public.users(id);
ALTER TABLE ONLY public.projects
    ADD CONSTRAINT fk_projects_deleted_by FOREIGN KEY (deleted_by) REFERENCES public.users(id);

ALTER TABLE ONLY public.managers
    ADD CONSTRAINT fk_managers_created_by FOREIGN KEY (created_by) REFERENCES public.users(id);
ALTER TABLE ONLY public.managers
    ADD CONSTRAINT fk_managers_updated_by FOREIGN KEY (updated_by) REFERENCES public.users(id);
ALTER TABLE ONLY public.managers
    ADD CONSTRAINT fk_managers_deleted_by FOREIGN KEY (deleted_by) REFERENCES public.users(id);

ALTER TABLE ONLY public.project_managers
    ADD CONSTRAINT fk_project_managers_project FOREIGN KEY (project_id) REFERENCES public.projects(id) ON DELETE CASCADE;
ALTER TABLE ONLY public.project_managers
    ADD CONSTRAINT fk_project_managers_manager FOREIGN KEY (manager_id) REFERENCES public.managers(id);
ALTER TABLE ONLY public.project_managers
    ADD CONSTRAINT fk_project_managers_created_by FOREIGN KEY (created_by) REFERENCES public.users(id);
ALTER TABLE ONLY public.project_managers
    ADD CONSTRAINT fk_project_managers_updated_by FOREIGN KEY (updated_by) REFERENCES public.users(id);
ALTER TABLE ONLY public.project_managers
    ADD CONSTRAINT fk_project_managers_deleted_by FOREIGN KEY (deleted_by) REFERENCES public.users(id);

ALTER TABLE ONLY public.lease_types
    ADD CONSTRAINT fk_lease_types_created_by FOREIGN KEY (created_by) REFERENCES public.users(id);
ALTER TABLE ONLY public.lease_types
    ADD CONSTRAINT fk_lease_types_updated_by FOREIGN KEY (updated_by) REFERENCES public.users(id);
ALTER TABLE ONLY public.lease_types
    ADD CONSTRAINT fk_lease_types_deleted_by FOREIGN KEY (deleted_by) REFERENCES public.users(id);

ALTER TABLE ONLY public.fields
    ADD CONSTRAINT fk_fields_project FOREIGN KEY (project_id) REFERENCES public.projects(id) ON DELETE CASCADE;
ALTER TABLE ONLY public.fields
    ADD CONSTRAINT fk_fields_lease_type FOREIGN KEY (lease_type_id) REFERENCES public.lease_types(id);
ALTER TABLE ONLY public.fields
    ADD CONSTRAINT fk_fields_created_by FOREIGN KEY (created_by) REFERENCES public.users(id);
ALTER TABLE ONLY public.fields
    ADD CONSTRAINT fk_fields_updated_by FOREIGN KEY (updated_by) REFERENCES public.users(id);
ALTER TABLE ONLY public.fields
    ADD CONSTRAINT fk_fields_deleted_by FOREIGN KEY (deleted_by) REFERENCES public.users(id);

ALTER TABLE ONLY public.lots
    ADD CONSTRAINT fk_lots_field FOREIGN KEY (field_id) REFERENCES public.fields(id) ON DELETE CASCADE;
ALTER TABLE ONLY public.lots
    ADD CONSTRAINT fk_lots_current_crop FOREIGN KEY (current_crop_id) REFERENCES public.crops(id);
ALTER TABLE ONLY public.lots
    ADD CONSTRAINT fk_lots_previous_crop FOREIGN KEY (previous_crop_id) REFERENCES public.crops(id);
ALTER TABLE ONLY public.lots
    ADD CONSTRAINT fk_lots_created_by FOREIGN KEY (created_by) REFERENCES public.users(id);
ALTER TABLE ONLY public.lots
    ADD CONSTRAINT fk_lots_updated_by FOREIGN KEY (updated_by) REFERENCES public.users(id);
ALTER TABLE ONLY public.lots
    ADD CONSTRAINT fk_lots_deleted_by FOREIGN KEY (deleted_by) REFERENCES public.users(id);

ALTER TABLE ONLY public.lot_dates
    ADD CONSTRAINT fk_lot_dates_lot FOREIGN KEY (lot_id) REFERENCES public.lots(id) ON DELETE CASCADE;
ALTER TABLE ONLY public.lot_dates
    ADD CONSTRAINT fk_lot_dates_created_by FOREIGN KEY (created_by) REFERENCES public.users(id);
ALTER TABLE ONLY public.lot_dates
    ADD CONSTRAINT fk_lot_dates_updated_by FOREIGN KEY (updated_by) REFERENCES public.users(id);
ALTER TABLE ONLY public.lot_dates
    ADD CONSTRAINT fk_lot_dates_deleted_by FOREIGN KEY (deleted_by) REFERENCES public.users(id);

ALTER TABLE ONLY public.crops
    ADD CONSTRAINT fk_crops_created_by FOREIGN KEY (created_by) REFERENCES public.users(id);
ALTER TABLE ONLY public.crops
    ADD CONSTRAINT fk_crops_updated_by FOREIGN KEY (updated_by) REFERENCES public.users(id);
ALTER TABLE ONLY public.crops
    ADD CONSTRAINT fk_crops_deleted_by FOREIGN KEY (deleted_by) REFERENCES public.users(id);

ALTER TABLE ONLY public.labor_categories
    ADD CONSTRAINT fk_labor_categories_type FOREIGN KEY (type_id) REFERENCES public.labor_types(id);

ALTER TABLE ONLY public.labors
    ADD CONSTRAINT fk_labors_project FOREIGN KEY (project_id) REFERENCES public.projects(id) ON UPDATE CASCADE ON DELETE RESTRICT;
ALTER TABLE ONLY public.labors
    ADD CONSTRAINT fk_labors_category FOREIGN KEY (category_id) REFERENCES public.categories(id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE ONLY public.workorders
    ADD CONSTRAINT fk_workorders_project FOREIGN KEY (project_id) REFERENCES public.projects(id) ON UPDATE CASCADE ON DELETE RESTRICT;
ALTER TABLE ONLY public.workorders
    ADD CONSTRAINT fk_workorders_field FOREIGN KEY (field_id) REFERENCES public.fields(id) ON UPDATE CASCADE ON DELETE RESTRICT;
ALTER TABLE ONLY public.workorders
    ADD CONSTRAINT fk_workorders_lot FOREIGN KEY (lot_id) REFERENCES public.lots(id) ON UPDATE CASCADE ON DELETE RESTRICT;
ALTER TABLE ONLY public.workorders
    ADD CONSTRAINT fk_workorders_crop FOREIGN KEY (crop_id) REFERENCES public.crops(id) ON UPDATE CASCADE ON DELETE RESTRICT;
ALTER TABLE ONLY public.workorders
    ADD CONSTRAINT fk_workorders_labor FOREIGN KEY (labor_id) REFERENCES public.labors(id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE ONLY public.workorder_items
    ADD CONSTRAINT fk_workorder_items_workorder FOREIGN KEY (workorder_id) REFERENCES public.workorders(id) ON UPDATE CASCADE ON DELETE CASCADE;
ALTER TABLE ONLY public.workorder_items
    ADD CONSTRAINT fk_workorder_items_supply FOREIGN KEY (supply_id) REFERENCES public.supplies(id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE ONLY public.categories
    ADD CONSTRAINT fk_categories_type FOREIGN KEY (type_id) REFERENCES public.types(id);

ALTER TABLE ONLY public.supplies
    ADD CONSTRAINT fk_supplies_category FOREIGN KEY (category_id) REFERENCES public.categories(id) ON UPDATE CASCADE ON DELETE RESTRICT;
ALTER TABLE ONLY public.supplies
    ADD CONSTRAINT fk_supplies_type FOREIGN KEY (type_id) REFERENCES public.types(id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE ONLY public.stocks
    ADD CONSTRAINT fk_stocks_project FOREIGN KEY (project_id) REFERENCES public.projects(id) ON UPDATE CASCADE ON DELETE RESTRICT;
ALTER TABLE ONLY public.stocks
    ADD CONSTRAINT fk_stocks_supply FOREIGN KEY (supply_id) REFERENCES public.supplies(id) ON UPDATE CASCADE ON DELETE RESTRICT;
ALTER TABLE ONLY public.stocks
    ADD CONSTRAINT fk_stocks_investor FOREIGN KEY (investor_id) REFERENCES public.investors(id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE ONLY public.supply_movements
    ADD CONSTRAINT fk_supply_movements_supply FOREIGN KEY (supply_id) REFERENCES public.supplies(id) ON UPDATE CASCADE ON DELETE RESTRICT;
ALTER TABLE ONLY public.supply_movements
    ADD CONSTRAINT fk_supply_movements_provider FOREIGN KEY (provider_id) REFERENCES public.providers(id) ON UPDATE CASCADE ON DELETE RESTRICT;
ALTER TABLE ONLY public.supply_movements
    ADD CONSTRAINT fk_supply_movements_investor FOREIGN KEY (investor_id) REFERENCES public.investors(id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE ONLY public.investors
    ADD CONSTRAINT fk_investors_created_by FOREIGN KEY (created_by) REFERENCES public.users(id);
ALTER TABLE ONLY public.investors
    ADD CONSTRAINT fk_investors_updated_by FOREIGN KEY (updated_by) REFERENCES public.users(id);
ALTER TABLE ONLY public.investors
    ADD CONSTRAINT fk_investors_deleted_by FOREIGN KEY (deleted_by) REFERENCES public.users(id);

ALTER TABLE ONLY public.project_investors
    ADD CONSTRAINT fk_project_investors_project FOREIGN KEY (project_id) REFERENCES public.projects(id) ON DELETE CASCADE;
ALTER TABLE ONLY public.project_investors
    ADD CONSTRAINT fk_project_investors_investor FOREIGN KEY (investor_id) REFERENCES public.investors(id);
ALTER TABLE ONLY public.project_investors
    ADD CONSTRAINT fk_project_investors_created_by FOREIGN KEY (created_by) REFERENCES public.users(id);
ALTER TABLE ONLY public.project_investors
    ADD CONSTRAINT fk_project_investors_updated_by FOREIGN KEY (updated_by) REFERENCES public.users(id);
ALTER TABLE ONLY public.project_investors
    ADD CONSTRAINT fk_project_investors_deleted_by FOREIGN KEY (deleted_by) REFERENCES public.users(id);

ALTER TABLE ONLY public.crop_commercializations
    ADD CONSTRAINT fk_crop_commercializations_project FOREIGN KEY (project_id) REFERENCES public.projects(id);
ALTER TABLE ONLY public.crop_commercializations
    ADD CONSTRAINT fk_crop_commercializations_crop FOREIGN KEY (crop_id) REFERENCES public.crops(id);

ALTER TABLE ONLY public.admin_cost_investors
    ADD CONSTRAINT fk_admin_cost_investors_project FOREIGN KEY (project_id) REFERENCES public.projects(id) ON DELETE CASCADE;
ALTER TABLE ONLY public.admin_cost_investors
    ADD CONSTRAINT fk_admin_cost_investors_investor FOREIGN KEY (investor_id) REFERENCES public.investors(id);

ALTER TABLE ONLY public.field_investors
    ADD CONSTRAINT fk_field_investors_field FOREIGN KEY (field_id) REFERENCES public.fields(id) ON DELETE CASCADE;
ALTER TABLE ONLY public.field_investors
    ADD CONSTRAINT fk_field_investors_investor FOREIGN KEY (investor_id) REFERENCES public.investors(id);

ALTER TABLE ONLY public.project_dollar_values
    ADD CONSTRAINT fk_project_dollar_values_project FOREIGN KEY (project_id) REFERENCES public.projects(id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE ONLY public.invoices
    ADD CONSTRAINT fk_invoices_work_order FOREIGN KEY (work_order_id) REFERENCES public.workorders(id) ON DELETE CASCADE;

-- INDEXES
CREATE INDEX idx_users_deleted_at ON public.users USING btree (deleted_at);
CREATE INDEX idx_business_parameters_category ON public.business_parameters USING btree (category);
CREATE INDEX idx_business_parameters_key ON public.business_parameters USING btree (key);
CREATE INDEX idx_fx_rates_effective_date ON public.fx_rates USING btree (effective_date);

CREATE INDEX idx_customers_created_by ON public.customers USING btree (created_by);
CREATE INDEX idx_customers_updated_by ON public.customers USING btree (updated_by);
CREATE INDEX idx_customers_deleted_by ON public.customers USING btree (deleted_by);

CREATE INDEX idx_campaigns_created_by ON public.campaigns USING btree (created_by);
CREATE INDEX idx_campaigns_updated_by ON public.campaigns USING btree (updated_by);
CREATE INDEX idx_campaigns_deleted_by ON public.campaigns USING btree (deleted_by);

CREATE INDEX idx_projects_notdel ON public.projects USING btree (id, admin_cost) WHERE (deleted_at IS NULL);
CREATE INDEX idx_projects_campaign_id ON public.projects USING btree (campaign_id);
CREATE INDEX idx_projects_customer_id ON public.projects USING btree (customer_id);
CREATE INDEX idx_projects_active ON public.projects USING btree (id) WHERE (deleted_at IS NULL);
CREATE INDEX idx_projects_created_by ON public.projects USING btree (created_by);
CREATE INDEX idx_projects_updated_by ON public.projects USING btree (updated_by);
CREATE INDEX idx_projects_deleted_by ON public.projects USING btree (deleted_by);

CREATE INDEX idx_managers_created_by ON public.managers USING btree (created_by);
CREATE INDEX idx_managers_updated_by ON public.managers USING btree (updated_by);
CREATE INDEX idx_managers_deleted_by ON public.managers USING btree (deleted_by);

CREATE INDEX idx_project_managers_manager_id ON public.project_managers USING btree (manager_id);
CREATE INDEX idx_project_managers_created_by ON public.project_managers USING btree (created_by);
CREATE INDEX idx_project_managers_updated_by ON public.project_managers USING btree (updated_by);
CREATE INDEX idx_project_managers_deleted_by ON public.project_managers USING btree (deleted_by);

CREATE INDEX idx_lease_types_created_by ON public.lease_types USING btree (created_by);
CREATE INDEX idx_lease_types_updated_by ON public.lease_types USING btree (updated_by);
CREATE INDEX idx_lease_types_deleted_by ON public.lease_types USING btree (deleted_by);

CREATE INDEX idx_fields_project_id ON public.fields USING btree (project_id);
CREATE INDEX idx_fields_notdel ON public.fields USING btree (id, project_id, lease_type_id, lease_type_value, lease_type_percent) WHERE (deleted_at IS NULL);
CREATE INDEX idx_fields_lease_type_id ON public.fields USING btree (lease_type_id);
CREATE INDEX idx_fields_created_by ON public.fields USING btree (created_by);
CREATE INDEX idx_fields_updated_by ON public.fields USING btree (updated_by);
CREATE INDEX idx_fields_deleted_by ON public.fields USING btree (deleted_by);

CREATE INDEX idx_lots_composite_notdel ON public.lots USING btree (field_id, current_crop_id, previous_crop_id, tons, hectares)
    WHERE ((deleted_at IS NULL) AND (hectares > (0)::numeric));
CREATE INDEX idx_lots_notdel ON public.lots USING btree (id, field_id, current_crop_id, previous_crop_id, tons, hectares)
    WHERE (deleted_at IS NULL);
CREATE INDEX idx_lots_current_crop_id ON public.lots USING btree (current_crop_id);
CREATE INDEX idx_lots_field_id ON public.lots USING btree (field_id);
CREATE INDEX idx_lots_previous_crop_id ON public.lots USING btree (previous_crop_id);
CREATE INDEX idx_lots_created_by ON public.lots USING btree (created_by);
CREATE INDEX idx_lots_updated_by ON public.lots USING btree (updated_by);
CREATE INDEX idx_lots_deleted_by ON public.lots USING btree (deleted_by);

CREATE INDEX idx_crops_notdel ON public.crops USING btree (id, name) WHERE (deleted_at IS NULL);
CREATE INDEX idx_crops_created_by ON public.crops USING btree (created_by);
CREATE INDEX idx_crops_updated_by ON public.crops USING btree (updated_by);
CREATE INDEX idx_crops_deleted_by ON public.crops USING btree (deleted_by);

CREATE INDEX idx_lot_dates_lot_id ON public.lot_dates USING btree (lot_id);
CREATE INDEX idx_lot_dates_created_by ON public.lot_dates USING btree (created_by);
CREATE INDEX idx_lot_dates_updated_by ON public.lot_dates USING btree (updated_by);
CREATE INDEX idx_lot_dates_deleted_by ON public.lot_dates USING btree (deleted_by);

CREATE INDEX idx_labor_categories_type_id ON public.labor_categories USING btree (type_id);

CREATE INDEX idx_labors_notdel ON public.labors USING btree (id, price) WHERE (deleted_at IS NULL);
CREATE INDEX idx_labors_category_id ON public.labors USING btree (category_id) WHERE (deleted_at IS NULL);
CREATE INDEX idx_labors_project_id ON public.labors USING btree (project_id) WHERE (deleted_at IS NULL);
CREATE INDEX idx_labors_harvest ON public.labors USING btree (id, category_id)
    WHERE ((deleted_at IS NULL) AND (category_id = 2));
CREATE INDEX idx_labors_notdel_price ON public.labors USING btree (id, category_id, price)
    WHERE (deleted_at IS NULL);
CREATE INDEX idx_labors_sowing ON public.labors USING btree (id, category_id)
    WHERE ((deleted_at IS NULL) AND (category_id = 1));

CREATE INDEX idx_workorders_grouping ON public.workorders USING btree (project_id, field_id, effective_area)
    WHERE ((deleted_at IS NULL) AND (effective_area > (0)::numeric));
CREATE INDEX idx_workorders_composite ON public.workorders USING btree (project_id, field_id, labor_id, effective_area)
    WHERE ((deleted_at IS NULL) AND (effective_area > (0)::numeric));
CREATE INDEX idx_workorders_metrics_notdel ON public.workorders USING btree (project_id, field_id, labor_id, effective_area)
    WHERE (deleted_at IS NULL);
CREATE INDEX idx_workorders_lot_composite ON public.workorders USING btree (lot_id, labor_id, effective_area, date)
    WHERE ((deleted_at IS NULL) AND (effective_area > (0)::numeric));
CREATE INDEX idx_workorders_lot_notdel ON public.workorders USING btree (lot_id, effective_area, date)
    WHERE (deleted_at IS NULL);
CREATE INDEX idx_workorders_field_id ON public.workorders USING btree (field_id);
CREATE INDEX idx_workorders_crop_id ON public.workorders USING btree (crop_id);
CREATE INDEX idx_workorders_labor_id ON public.workorders USING btree (labor_id);

CREATE INDEX idx_workorder_items_notdel ON public.workorder_items USING btree (workorder_id, supply_id)
    WHERE (deleted_at IS NULL);
CREATE INDEX idx_workorder_items_supply_notdel ON public.workorder_items USING btree (workorder_id, supply_id, final_dose)
    WHERE (deleted_at IS NULL);
CREATE INDEX idx_workorder_items_v2_notdel ON public.workorder_items USING btree (workorder_id, supply_id, total_used, final_dose)
    WHERE (deleted_at IS NULL);
CREATE INDEX idx_workorder_items_supply_id ON public.workorder_items USING btree (supply_id)
    WHERE (deleted_at IS NULL);
CREATE INDEX idx_workorder_items_workorder_id ON public.workorder_items USING btree (workorder_id)
    WHERE (deleted_at IS NULL);

CREATE INDEX idx_supplies_notdel ON public.supplies USING btree (id, price) WHERE (deleted_at IS NULL);
CREATE INDEX idx_supplies_units_notdel ON public.supplies USING btree (id, price, unit_id) WHERE (deleted_at IS NULL);
CREATE INDEX idx_supplies_category_id ON public.supplies USING btree (category_id);
CREATE INDEX idx_supplies_type_id ON public.supplies USING btree (type_id);

CREATE INDEX idx_categories_type_id ON public.categories USING btree (type_id);

CREATE INDEX idx_stocks_project_id ON public.stocks USING btree (project_id);
CREATE INDEX idx_stocks_supply_id ON public.stocks USING btree (supply_id);
CREATE INDEX idx_stocks_investor_id ON public.stocks USING btree (investor_id);

CREATE INDEX idx_supply_movements_supply_id ON public.supply_movements USING btree (supply_id);
CREATE INDEX idx_supply_movements_provider_id ON public.supply_movements USING btree (provider_id);
CREATE INDEX idx_supply_movements_investor_id ON public.supply_movements USING btree (investor_id);

CREATE INDEX idx_crop_commercializations_crop_id ON public.crop_commercializations USING btree (crop_id) WHERE (deleted_at IS NULL);
CREATE INDEX idx_crop_commercializations_deleted_at ON public.crop_commercializations USING btree (deleted_at);
CREATE INDEX idx_crop_commercializations_project_id ON public.crop_commercializations USING btree (project_id);
CREATE INDEX idx_crop_commercializations_notdel ON public.crop_commercializations USING btree (project_id, crop_id, net_price)
    WHERE (deleted_at IS NULL);

CREATE INDEX idx_project_dollar_values_project_id ON public.project_dollar_values USING btree (project_id) WHERE (deleted_at IS NULL);

CREATE INDEX idx_investors_created_by ON public.investors USING btree (created_by);
CREATE INDEX idx_investors_updated_by ON public.investors USING btree (updated_by);
CREATE INDEX idx_investors_deleted_by ON public.investors USING btree (deleted_by);

CREATE INDEX idx_project_investors_investor_id ON public.project_investors USING btree (investor_id);
CREATE INDEX idx_project_investors_created_by ON public.project_investors USING btree (created_by);
CREATE INDEX idx_project_investors_updated_by ON public.project_investors USING btree (updated_by);
CREATE INDEX idx_project_investors_deleted_by ON public.project_investors USING btree (deleted_by);

CREATE INDEX idx_admin_cost_investors_investor_id ON public.admin_cost_investors USING btree (investor_id);
CREATE INDEX idx_field_investors_investor_id ON public.field_investors USING btree (investor_id);

COMMIT;
