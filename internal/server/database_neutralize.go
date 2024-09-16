package server

func (db *Database) RemoveEnterpriseCode() error {
	// -- remove the enterprise code, report.url and web.base.url
	_, err := db.Exec("delete from ir_config_parameter where key in ('database.enterprise_code', 'report.url', 'web.base.url.freeze')")
	if err != nil {
		return err
	}
	return nil
}

func (db *Database) ChangeDBUUID() error {
	// reset db uuid
	_, err := db.Exec(`update ir_config_parameter set value=(select gen_random_uuid()) where key = 'database.uuid'`)
	if err != nil {
		return err
	}
	return nil
}

func (db *Database) UpdateDatabaseExpirationDate() error {
	// update expiration date
	_, err := db.Exec(`insert into ir_config_parameter
		(key,value,create_uid,create_date,write_uid,write_date)
		values
		('database.expiration_date',(current_date+'3 months'::interval)::timestamp,1,
		current_timestamp,1,current_timestamp)
		on conflict (key)
		do UPDATE set value = (current_date+'3 months'::interval)::timestamp;`)
	if err != nil {
		return err
	}
	return nil
}

func (db *Database) DisableBankSync() error {
	// disable bank synchronisation links
	_, err := db.Exec(`UPDATE account_online_link SET provider_data = '', client_id = 'duplicate';`)
	if err != nil {
		return err
	}
	return nil
}

func (db *Database) DisableFetchmail() error {
	// deactivate fetchmail server
	_, err := db.Exec("UPDATE fetchmail_server SET active = false;")
	if err != nil {
		return err
	}
	return nil
}

func (db *Database) DeactivateMailServers() error {
	// deactivate mail servers but activate default "localhost" mail server
	_, err := db.Exec(`DO $$
		        BEGIN
		            UPDATE ir_mail_server SET active = 'f';
		            IF EXISTS (SELECT 1 FROM ir_module_module WHERE name='mail' and state IN ('installed', 'to upgrade', 'to remove')) THEN
		                UPDATE mail_template SET mail_server_id = NULL;
		            END IF;
		        EXCEPTION
		            WHEN undefined_table OR undefined_column THEN
		        END;
		    $$;`)
	if err != nil {
		return err
	}
	return nil
}

func (db *Database) DeactivateCrons() error {
	// deactivate crons
	_, err := db.Exec("UPDATE ir_cron SET active = 'f';")
	if err != nil {
		return err
	}
	_, err = db.Exec(`UPDATE ir_cron SET active = 't' WHERE id IN (SELECT res_id FROM ir_model_data WHERE name = 'autovacuum_job' AND module = 'base');`)
	if err != nil {
		return err
	}
	return nil
}

func (db *Database) ActivateModuleUpdateNotificationCron() error {
	// activating module update notification cron
	_, err := db.Exec(`UPDATE ir_cron SET active = 't' WHERE id IN (SELECT res_id FROM ir_model_data WHERE name = 'ir_cron_module_update_notification' AND module = 'mail');`)
	if err != nil {
		return err
	}
	return nil
}

func (db *Database) RemoveIRLogging() error {
	// remove platform ir_logging
	_, err := db.Exec("DELETE FROM ir_logging WHERE func = 'odoo.sh';")
	if err != nil {
		return err
	}
	return nil
}

func (db *Database) DisableProdDeliveryCarriers() error {
	// disable prod environment in all delivery carriers
	_, err := db.Exec("UPDATE delivery_carrier SET prod_environment = false;")
	if err != nil {
		return err
	}
	return nil
}

func (db *Database) DisableDeliveryCarriers() error {
	// disable delivery carriers from external providers
	_, err := db.Exec("UPDATE delivery_carrier SET active = false WHERE delivery_type NOT IN ('fixed', 'base_on_rule');")
	if err != nil {
		return err
	}
	return nil
}

func (db *Database) DisableIAPAccount() error {
	// disable iap account
	_, err := db.Exec(`UPDATE iap_account SET account_token = REGEXP_REPLACE(account_token, '(\+.*)?$', '+disabled');`)
	if err != nil {
		return err
	}
	return nil
}

func (db *Database) DisableMailTemplate() error {
	// deactivate mail template
	_, err := db.Exec("UPDATE mail_template SET mail_server_id = NULL;")
	if err != nil {
		return err
	}
	return nil
}

func (db *Database) DisablePaymentGeneric() error {
	// disable generic payment provider
	_, err := db.Exec("UPDATE payment_provider SET state = 'disabled' WHERE state NOT IN ('test', 'disabled');")
	if err != nil {
		return err
	}
	return nil
}

func (db *Database) DeleteWebsiteDomains() error {
	// delete domains on websites
	_, err := db.Exec("UPDATE website SET domain = NULL;")
	if err != nil {
		return err
	}
	return nil
}

func (db *Database) DisableCDN() error {
	// disable cdn
	_, err := db.Exec("UPDATE website SET cdn_activated = false;")
	if err != nil {
		return err
	}
	return nil
}

func (db *Database) DeleteOCNProjectUUID() error {
	// delete odoo_ocn.project_id and ocn.uuid
	_, err := db.Exec("DELETE FROM ir_config_parameter WHERE key IN ('odoo_ocn.project_id', 'ocn.uuid');")
	if err != nil {
		return err
	}
	return nil
}

func (db *Database) RemoveFacebookTokens() error {
	// delete Facebook Access Tokens
	_, err := db.Exec("UPDATE social_account SET facebook_account_id = NULL, facebook_access_token = NULL;")
	if err != nil {
		return err
	}
	return nil
}

func (db *Database) RemoveInstagramTokens() error {
	// delete Instagram Access Tokens
	_, err := db.Exec("UPDATE social_account SET instagram_account_id = NULL, instagram_facebook_account_id = NULL, instagram_access_token = NULL;")
	if err != nil {
		return err
	}
	return nil
}

func (db *Database) RemoveLinkedInTokens() error {
	// delete LinkedIn Access Tokens
	_, err := db.Exec("UPDATE social_account SET linkedin_account_urn = NULL, linkedin_access_token = NULL;")
	if err != nil {
		return err
	}
	return nil
}

func (db *Database) RemoveTwitterTokens() error {
	// delete Twitter Access Tokens
	_, err := db.Exec("UPDATE social_account SET twitter_user_id = NULL, twitter_oauth_token = NULL, twitter_oauth_token_secret = NULL;")
	if err != nil {
		return err
	}
	return nil
}

func (db *Database) RemoveYoutubeTokens() error {
	// delete Youtube Access Tokens
	_, err := db.Exec("UPDATE social_account SET youtube_channel_id = NULL, youtube_access_token = NULL, youtube_refresh_token = NULL, youtube_token_expiration_date = NULL, youtube_upload_playlist_id = NULL;")
	if err != nil {
		return err
	}
	return nil
}

func (db *Database) UnsetFirebase() error {
	// Unset Firebase configuration within website
	_, err := db.Exec("UPDATE website SET firebase_enable_push_notifications = false, firebase_use_own_account = false, firebase_project_id = NULL, firebase_web_api_key = NULL, firebase_push_certificate_key = NULL, firebase_sender_id = NULL;")
	if err != nil {
		return err
	}
	return nil
}

func (db *Database) RemoveMapBoxToken() error {
	// Remove Map Box Token as it's only valid per DB url
	_, err := db.Exec("DELETE FROM ir_config_parameter WHERE key = 'web_map.token_map_box';")
	if err != nil {
		return err
	}
	return nil
}

func (db *Database) ActivateNeutralizationWatermarks() error {
	// activate neutralization watermarks banner
	_, err := db.Exec("UPDATE ir_ui_view SET active = true WHERE key = 'web.neutralize_banner';")
	if err != nil {
		return err
	}
	// activate neutralization watermarks ribbon
	_, err = db.Exec("UPDATE ir_ui_view SET active = true WHERE key = 'website.neutralize_ribbon';")
	if err != nil {
		return err
	}
	return nil
}
