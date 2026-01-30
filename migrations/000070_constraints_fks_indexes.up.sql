-- ========================================
-- MIGRATION 000070 CONSTRAINTS FKS INDEXES (UP)
-- ========================================
-- Nota: SQL en inglés, comentarios en español

BEGIN;

ALTER TABLE ONLY public.users
    ADD CONSTRAINT uni_users_username UNIQUE (username);

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_username_key UNIQUE (username);

ALTER TABLE ONLY public.business_parameters
    ADD CONSTRAINT business_parameters_key_key UNIQUE (key);

ALTER TABLE ONLY public.business_parameters
    ADD CONSTRAINT business_parameters_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.fx_rates
    ADD CONSTRAINT fx_rates_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.customers
    ADD CONSTRAINT customers_name_key UNIQUE (name);

ALTER TABLE ONLY public.customers
    ADD CONSTRAINT customers_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.campaigns
    ADD CONSTRAINT campaigns_name_key UNIQUE (name);

ALTER TABLE ONLY public.campaigns
    ADD CONSTRAINT campaigns_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.projects
    ADD CONSTRAINT projects_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.managers
    ADD CONSTRAINT managers_name_key UNIQUE (name);

ALTER TABLE ONLY public.managers
    ADD CONSTRAINT managers_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.project_managers
    ADD CONSTRAINT project_managers_pkey PRIMARY KEY (project_id, manager_id);

ALTER TABLE ONLY public.lease_types
    ADD CONSTRAINT lease_types_name_key UNIQUE (name);

ALTER TABLE ONLY public.lease_types
    ADD CONSTRAINT lease_types_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.fields
    ADD CONSTRAINT fields_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.lots
    ADD CONSTRAINT lots_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.lot_dates
    ADD CONSTRAINT lot_dates_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.crops
    ADD CONSTRAINT crops_name_key UNIQUE (name);

ALTER TABLE ONLY public.crops
    ADD CONSTRAINT crops_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.labor_types
    ADD CONSTRAINT labor_types_name_key UNIQUE (name);

ALTER TABLE ONLY public.labor_types
    ADD CONSTRAINT labor_types_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.labor_categories
    ADD CONSTRAINT labor_categories_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.labors
    ADD CONSTRAINT labors_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.workorders
    ADD CONSTRAINT workorders_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.workorder_items
    ADD CONSTRAINT workorder_items_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.types
    ADD CONSTRAINT types_name_key UNIQUE (name);

ALTER TABLE ONLY public.types
    ADD CONSTRAINT types_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.categories
    ADD CONSTRAINT categories_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.supplies
    ADD CONSTRAINT supplies_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.stocks
    ADD CONSTRAINT stocks_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.supply_movements
    ADD CONSTRAINT supply_movements_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.providers
    ADD CONSTRAINT providers_name_key UNIQUE (name);

ALTER TABLE ONLY public.providers
    ADD CONSTRAINT providers_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.investors
    ADD CONSTRAINT investors_name_key UNIQUE (name);

ALTER TABLE ONLY public.investors
    ADD CONSTRAINT investors_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.project_investors
    ADD CONSTRAINT project_investors_pkey PRIMARY KEY (project_id, investor_id);

ALTER TABLE ONLY public.crop_commercializations
    ADD CONSTRAINT crop_commercializations_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.admin_cost_investors
    ADD CONSTRAINT admin_cost_investors_pkey PRIMARY KEY (project_id, investor_id);

ALTER TABLE ONLY public.field_investors
    ADD CONSTRAINT field_investors_pkey PRIMARY KEY (field_id, investor_id);

ALTER TABLE ONLY public.project_dollar_values
    ADD CONSTRAINT project_dollar_values_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.project_dollar_values
    ADD CONSTRAINT project_dollar_values_project_id_year_month_key UNIQUE (project_id, year, month);

ALTER TABLE ONLY public.invoices
    ADD CONSTRAINT invoices_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.invoices
    ADD CONSTRAINT invoices_work_order_id_key UNIQUE (work_order_id);

ALTER TABLE ONLY public.customers
    ADD CONSTRAINT fk_customers_created_by FOREIGN KEY (created_by) REFERENCES public.users(id);

ALTER TABLE ONLY public.customers
    ADD CONSTRAINT fk_customers_deleted_by FOREIGN KEY (deleted_by) REFERENCES public.users(id);

ALTER TABLE ONLY public.customers
    ADD CONSTRAINT fk_customers_updated_by FOREIGN KEY (updated_by) REFERENCES public.users(id);

ALTER TABLE ONLY public.campaigns
    ADD CONSTRAINT fk_campaigns_created_by FOREIGN KEY (created_by) REFERENCES public.users(id);

ALTER TABLE ONLY public.campaigns
    ADD CONSTRAINT fk_campaigns_deleted_by FOREIGN KEY (deleted_by) REFERENCES public.users(id);

ALTER TABLE ONLY public.campaigns
    ADD CONSTRAINT fk_campaigns_updated_by FOREIGN KEY (updated_by) REFERENCES public.users(id);

ALTER TABLE ONLY public.projects
    ADD CONSTRAINT fk_projects_campaign FOREIGN KEY (campaign_id) REFERENCES public.campaigns(id);

ALTER TABLE ONLY public.projects
    ADD CONSTRAINT fk_projects_created_by FOREIGN KEY (created_by) REFERENCES public.users(id);

ALTER TABLE ONLY public.projects
    ADD CONSTRAINT fk_projects_customer FOREIGN KEY (customer_id) REFERENCES public.customers(id);

ALTER TABLE ONLY public.projects
    ADD CONSTRAINT fk_projects_deleted_by FOREIGN KEY (deleted_by) REFERENCES public.users(id);

ALTER TABLE ONLY public.projects
    ADD CONSTRAINT fk_projects_updated_by FOREIGN KEY (updated_by) REFERENCES public.users(id);

ALTER TABLE ONLY public.managers
    ADD CONSTRAINT fk_managers_created_by FOREIGN KEY (created_by) REFERENCES public.users(id);

ALTER TABLE ONLY public.managers
    ADD CONSTRAINT fk_managers_deleted_by FOREIGN KEY (deleted_by) REFERENCES public.users(id);

ALTER TABLE ONLY public.managers
    ADD CONSTRAINT fk_managers_updated_by FOREIGN KEY (updated_by) REFERENCES public.users(id);

ALTER TABLE ONLY public.project_managers
    ADD CONSTRAINT fk_project_managers_created_by FOREIGN KEY (created_by) REFERENCES public.users(id);

ALTER TABLE ONLY public.project_managers
    ADD CONSTRAINT fk_project_managers_deleted_by FOREIGN KEY (deleted_by) REFERENCES public.users(id);

ALTER TABLE ONLY public.project_managers
    ADD CONSTRAINT fk_project_managers_manager FOREIGN KEY (manager_id) REFERENCES public.managers(id);

ALTER TABLE ONLY public.project_managers
    ADD CONSTRAINT fk_project_managers_project FOREIGN KEY (project_id) REFERENCES public.projects(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.project_managers
    ADD CONSTRAINT fk_project_managers_updated_by FOREIGN KEY (updated_by) REFERENCES public.users(id);

ALTER TABLE ONLY public.lease_types
    ADD CONSTRAINT fk_lease_types_created_by FOREIGN KEY (created_by) REFERENCES public.users(id);

ALTER TABLE ONLY public.lease_types
    ADD CONSTRAINT fk_lease_types_deleted_by FOREIGN KEY (deleted_by) REFERENCES public.users(id);

ALTER TABLE ONLY public.lease_types
    ADD CONSTRAINT fk_lease_types_updated_by FOREIGN KEY (updated_by) REFERENCES public.users(id);

ALTER TABLE ONLY public.fields
    ADD CONSTRAINT fk_fields_created_by FOREIGN KEY (created_by) REFERENCES public.users(id);

ALTER TABLE ONLY public.fields
    ADD CONSTRAINT fk_fields_deleted_by FOREIGN KEY (deleted_by) REFERENCES public.users(id);

ALTER TABLE ONLY public.fields
    ADD CONSTRAINT fk_fields_project FOREIGN KEY (project_id) REFERENCES public.projects(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.fields
    ADD CONSTRAINT fk_fields_updated_by FOREIGN KEY (updated_by) REFERENCES public.users(id);

ALTER TABLE ONLY public.fields
    ADD CONSTRAINT fk_lease_type FOREIGN KEY (lease_type_id) REFERENCES public.lease_types(id);

ALTER TABLE ONLY public.lots
    ADD CONSTRAINT fk_lots_created_by FOREIGN KEY (created_by) REFERENCES public.users(id);

ALTER TABLE ONLY public.lot_dates
    ADD CONSTRAINT fk_lot_dates_created_by FOREIGN KEY (created_by) REFERENCES public.users(id);

ALTER TABLE ONLY public.lots
    ADD CONSTRAINT fk_lots_current_crop FOREIGN KEY (current_crop_id) REFERENCES public.crops(id);

ALTER TABLE ONLY public.lots
    ADD CONSTRAINT fk_lots_deleted_by FOREIGN KEY (deleted_by) REFERENCES public.users(id);

ALTER TABLE ONLY public.lot_dates
    ADD CONSTRAINT fk_lot_dates_deleted_by FOREIGN KEY (deleted_by) REFERENCES public.users(id);

ALTER TABLE ONLY public.lots
    ADD CONSTRAINT fk_lots_field FOREIGN KEY (field_id) REFERENCES public.fields(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.lot_dates
    ADD CONSTRAINT fk_lot_dates_lot FOREIGN KEY (lot_id) REFERENCES public.lots(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.lots
    ADD CONSTRAINT fk_lots_previous_crop FOREIGN KEY (previous_crop_id) REFERENCES public.crops(id);

ALTER TABLE ONLY public.lots
    ADD CONSTRAINT fk_lots_updated_by FOREIGN KEY (updated_by) REFERENCES public.users(id);

ALTER TABLE ONLY public.lot_dates
    ADD CONSTRAINT fk_lot_dates_updated_by FOREIGN KEY (updated_by) REFERENCES public.users(id);

ALTER TABLE ONLY public.crops
    ADD CONSTRAINT fk_crops_created_by FOREIGN KEY (created_by) REFERENCES public.users(id);

ALTER TABLE ONLY public.crops
    ADD CONSTRAINT fk_crops_deleted_by FOREIGN KEY (deleted_by) REFERENCES public.users(id);

ALTER TABLE ONLY public.crops
    ADD CONSTRAINT fk_crops_updated_by FOREIGN KEY (updated_by) REFERENCES public.users(id);

ALTER TABLE ONLY public.labor_categories
    ADD CONSTRAINT labor_categories_type_id_fkey FOREIGN KEY (type_id) REFERENCES public.labor_types(id);

ALTER TABLE ONLY public.labors
    ADD CONSTRAINT fk_category FOREIGN KEY (category_id) REFERENCES public.categories(id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE ONLY public.labors
    ADD CONSTRAINT fk_project FOREIGN KEY (project_id) REFERENCES public.projects(id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE ONLY public.workorders
    ADD CONSTRAINT workorders_crop_id_fkey FOREIGN KEY (crop_id) REFERENCES public.crops(id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE ONLY public.workorders
    ADD CONSTRAINT workorders_field_id_fkey FOREIGN KEY (field_id) REFERENCES public.fields(id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE ONLY public.workorders
    ADD CONSTRAINT workorders_labor_id_fkey FOREIGN KEY (labor_id) REFERENCES public.labors(id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE ONLY public.workorders
    ADD CONSTRAINT workorders_lot_id_fkey FOREIGN KEY (lot_id) REFERENCES public.lots(id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE ONLY public.workorders
    ADD CONSTRAINT workorders_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE ONLY public.workorder_items
    ADD CONSTRAINT workorder_items_supply_id_fkey FOREIGN KEY (supply_id) REFERENCES public.supplies(id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE ONLY public.workorder_items
    ADD CONSTRAINT workorder_items_workorder_id_fkey FOREIGN KEY (workorder_id) REFERENCES public.workorders(id) ON UPDATE CASCADE ON DELETE CASCADE;

ALTER TABLE ONLY public.categories
    ADD CONSTRAINT categories_type_id_fkey FOREIGN KEY (type_id) REFERENCES public.types(id);

ALTER TABLE ONLY public.supplies
    ADD CONSTRAINT fk_category FOREIGN KEY (category_id) REFERENCES public.categories(id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE ONLY public.supplies
    ADD CONSTRAINT fk_type FOREIGN KEY (type_id) REFERENCES public.types(id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE ONLY public.stocks
    ADD CONSTRAINT fk_investor FOREIGN KEY (investor_id) REFERENCES public.investors(id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE ONLY public.stocks
    ADD CONSTRAINT fk_project FOREIGN KEY (project_id) REFERENCES public.projects(id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE ONLY public.stocks
    ADD CONSTRAINT fk_supply FOREIGN KEY (supply_id) REFERENCES public.supplies(id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE ONLY public.supply_movements
    ADD CONSTRAINT fk_investor FOREIGN KEY (investor_id) REFERENCES public.investors(id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE ONLY public.supply_movements
    ADD CONSTRAINT fk_provider FOREIGN KEY (provider_id) REFERENCES public.providers(id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE ONLY public.supply_movements
    ADD CONSTRAINT fk_supply FOREIGN KEY (supply_id) REFERENCES public.supplies(id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE ONLY public.investors
    ADD CONSTRAINT fk_investors_created_by FOREIGN KEY (created_by) REFERENCES public.users(id);

ALTER TABLE ONLY public.investors
    ADD CONSTRAINT fk_investors_deleted_by FOREIGN KEY (deleted_by) REFERENCES public.users(id);

ALTER TABLE ONLY public.investors
    ADD CONSTRAINT fk_investors_updated_by FOREIGN KEY (updated_by) REFERENCES public.users(id);

ALTER TABLE ONLY public.project_investors
    ADD CONSTRAINT fk_project_investors_created_by FOREIGN KEY (created_by) REFERENCES public.users(id);

ALTER TABLE ONLY public.project_investors
    ADD CONSTRAINT fk_project_investors_deleted_by FOREIGN KEY (deleted_by) REFERENCES public.users(id);

ALTER TABLE ONLY public.project_investors
    ADD CONSTRAINT fk_project_investors_investor FOREIGN KEY (investor_id) REFERENCES public.investors(id);

ALTER TABLE ONLY public.project_investors
    ADD CONSTRAINT fk_project_investors_project FOREIGN KEY (project_id) REFERENCES public.projects(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.project_investors
    ADD CONSTRAINT fk_project_investors_updated_by FOREIGN KEY (updated_by) REFERENCES public.users(id);

ALTER TABLE ONLY public.crop_commercializations
    ADD CONSTRAINT fk_crop FOREIGN KEY (crop_id) REFERENCES public.crops(id);

ALTER TABLE ONLY public.crop_commercializations
    ADD CONSTRAINT fk_project FOREIGN KEY (project_id) REFERENCES public.projects(id);

ALTER TABLE ONLY public.admin_cost_investors
    ADD CONSTRAINT fk_admin_cost_investors_investor FOREIGN KEY (investor_id) REFERENCES public.investors(id);

ALTER TABLE ONLY public.admin_cost_investors
    ADD CONSTRAINT fk_admin_cost_investors_project FOREIGN KEY (project_id) REFERENCES public.projects(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.field_investors
    ADD CONSTRAINT fk_field_investors_field FOREIGN KEY (field_id) REFERENCES public.fields(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.field_investors
    ADD CONSTRAINT fk_field_investors_investor FOREIGN KEY (investor_id) REFERENCES public.investors(id);

ALTER TABLE ONLY public.project_dollar_values
    ADD CONSTRAINT fk_project FOREIGN KEY (project_id) REFERENCES public.projects(id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE ONLY public.invoices
    ADD CONSTRAINT fk_invoices_work_order FOREIGN KEY (work_order_id) REFERENCES public.workorders(id) ON DELETE CASCADE;

CREATE INDEX idx_users_deleted_at ON public.users USING btree (deleted_at);

CREATE INDEX idx_business_parameters_category ON public.business_parameters USING btree (category);

CREATE INDEX idx_business_parameters_key ON public.business_parameters USING btree (key);

CREATE INDEX idx_fx_rates_effective_date ON public.fx_rates USING btree (effective_date);

CREATE INDEX idx_lot_table_projects_notdel ON public.projects USING btree (id, admin_cost) WHERE (deleted_at IS NULL);

CREATE INDEX idx_projects_campaign_id ON public.projects USING btree (campaign_id);

CREATE INDEX idx_projects_customer_id ON public.projects USING btree (customer_id);

CREATE INDEX idx_projects_id ON public.projects USING btree (id) WHERE (deleted_at IS NULL);

CREATE INDEX idx_fields_project_id ON public.fields USING btree (project_id);

CREATE INDEX idx_lot_table_fields_notdel ON public.fields USING btree (id, project_id, lease_type_id, lease_type_value, lease_type_percent) WHERE (deleted_at IS NULL);

CREATE INDEX idx_lot_table_lots_composite ON public.lots USING btree (field_id, current_crop_id, previous_crop_id, tons, hectares) WHERE ((deleted_at IS NULL) AND (hectares > (0)::numeric));

CREATE INDEX idx_lot_table_lots_notdel ON public.lots USING btree (id, field_id, current_crop_id, previous_crop_id, tons, hectares) WHERE (deleted_at IS NULL);

CREATE INDEX idx_lots_current_crop_id ON public.lots USING btree (current_crop_id);

CREATE INDEX idx_lots_field_id ON public.lots USING btree (field_id);

CREATE INDEX idx_lots_previous_crop_id ON public.lots USING btree (previous_crop_id);

CREATE INDEX idx_lot_table_crops_notdel ON public.crops USING btree (id, name) WHERE (deleted_at IS NULL);

CREATE INDEX idx_labor_labors_notdel ON public.labors USING btree (id, price) WHERE (deleted_at IS NULL);

CREATE INDEX idx_labors_category_id ON public.labors USING btree (category_id) WHERE (deleted_at IS NULL);

CREATE INDEX idx_labors_project_id ON public.labors USING btree (project_id) WHERE (deleted_at IS NULL);

CREATE INDEX idx_lot_table_labors_harvest ON public.labors USING btree (id, category_id) WHERE ((deleted_at IS NULL) AND (category_id = 2));

CREATE INDEX idx_lot_table_labors_notdel ON public.labors USING btree (id, category_id, price) WHERE (deleted_at IS NULL);

CREATE INDEX idx_lot_table_labors_sowing ON public.labors USING btree (id, category_id) WHERE ((deleted_at IS NULL) AND (category_id = 1));

CREATE INDEX idx_labor_grouping ON public.workorders USING btree (project_id, field_id, effective_area) WHERE ((deleted_at IS NULL) AND (effective_area > (0)::numeric));

CREATE INDEX idx_labor_workorders_composite ON public.workorders USING btree (project_id, field_id, labor_id, effective_area) WHERE ((deleted_at IS NULL) AND (effective_area > (0)::numeric));

CREATE INDEX idx_labor_workorders_metrics_notdel ON public.workorders USING btree (project_id, field_id, labor_id, effective_area) WHERE (deleted_at IS NULL);

CREATE INDEX idx_labor_workorders_metrics_v2 ON public.workorders USING btree (project_id, field_id, labor_id, effective_area) WHERE ((deleted_at IS NULL) AND (effective_area > (0)::numeric));

CREATE INDEX idx_lot_table_workorders_composite ON public.workorders USING btree (lot_id, labor_id, effective_area, date) WHERE ((deleted_at IS NULL) AND (effective_area > (0)::numeric));

CREATE INDEX idx_lot_table_workorders_notdel ON public.workorders USING btree (lot_id, effective_area, date) WHERE (deleted_at IS NULL);

CREATE INDEX idx_labor_workorder_items_notdel ON public.workorder_items USING btree (workorder_id, supply_id) WHERE (deleted_at IS NULL);

CREATE INDEX idx_labor_workorder_items_supply ON public.workorder_items USING btree (workorder_id, supply_id, final_dose) WHERE (deleted_at IS NULL);

CREATE INDEX idx_labor_workorder_items_v2 ON public.workorder_items USING btree (workorder_id, supply_id, total_used, final_dose) WHERE (deleted_at IS NULL);

CREATE INDEX idx_lot_table_workorder_items_notdel ON public.workorder_items USING btree (workorder_id, supply_id, final_dose) WHERE (deleted_at IS NULL);

CREATE INDEX idx_workorder_items_supply_id ON public.workorder_items USING btree (supply_id) WHERE (deleted_at IS NULL);

CREATE INDEX idx_workorder_items_workorder_id ON public.workorder_items USING btree (workorder_id) WHERE (deleted_at IS NULL);

CREATE INDEX idx_labor_supplies_notdel ON public.supplies USING btree (id, price) WHERE (deleted_at IS NULL);

CREATE INDEX idx_labor_supplies_units_v2 ON public.supplies USING btree (id, price, unit_id) WHERE (deleted_at IS NULL);

CREATE INDEX idx_lot_table_supplies_notdel ON public.supplies USING btree (id, price) WHERE (deleted_at IS NULL);

CREATE INDEX idx_crop_commercializations_crop_id ON public.crop_commercializations USING btree (crop_id) WHERE (deleted_at IS NULL);

CREATE INDEX idx_crop_commercializations_deleted_at ON public.crop_commercializations USING btree (deleted_at);

CREATE INDEX idx_crop_commercializations_project_id ON public.crop_commercializations USING btree (project_id);

CREATE INDEX idx_lot_table_crop_commercializations_notdel ON public.crop_commercializations USING btree (project_id, crop_id, net_price) WHERE (deleted_at IS NULL);

CREATE INDEX idx_project_dollar_values_project_id ON public.project_dollar_values USING btree (project_id) WHERE (deleted_at IS NULL);

COMMIT;
