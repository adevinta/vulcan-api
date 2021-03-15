-- Add audit table --

CREATE TABLE audit (
    id          serial,
    date        timestamp DEFAULT now(),
    schema      text,
    tablename   text,
    operation   text,
    who         text DEFAULT current_user,
    new_val     json,
    old_val     json
);

-- ------------------
CREATE FUNCTION audit_function() RETURNS trigger AS $$
    BEGIN
        IF TG_OP = 'INSERT' THEN
            INSERT INTO audit (tablename, schema, operation, new_val)
                 VALUES (TG_RELNAME, TG_TABLE_SCHEMA, TG_OP, row_to_json(NEW));
            RETURN NEW;
        ELSIF TG_OP = 'UPDATE' THEN
            INSERT INTO audit (tablename, schema, operation, new_val, old_val)
                 VALUES (TG_RELNAME, TG_TABLE_SCHEMA, TG_OP, row_to_json(NEW), row_to_json(OLD));
            RETURN NEW;
        ELSIF TG_OP = 'DELETE' THEN
            INSERT INTO audit (tablename, schema, operation, old_val)
                 VALUES (TG_RELNAME, TG_TABLE_SCHEMA, TG_OP, row_to_json(OLD));
            RETURN OLD;
        END IF;
    END;
$$ LANGUAGE 'plpgsql' SECURITY DEFINER;

CREATE TRIGGER audit_trigger_asset_group BEFORE INSERT OR UPDATE OR DELETE ON asset_group FOR EACH ROW EXECUTE PROCEDURE audit_function();
CREATE TRIGGER audit_trigger_asset_types BEFORE INSERT OR UPDATE OR DELETE ON asset_types FOR EACH ROW EXECUTE PROCEDURE audit_function();
CREATE TRIGGER audit_trigger_assets BEFORE INSERT OR UPDATE OR DELETE ON assets FOR EACH ROW EXECUTE PROCEDURE audit_function();
CREATE TRIGGER audit_trigger_checktype_settings BEFORE INSERT OR UPDATE OR DELETE ON checktype_settings FOR EACH ROW EXECUTE PROCEDURE audit_function();
CREATE TRIGGER audit_trigger_global_programs_metadata BEFORE INSERT OR UPDATE OR DELETE ON global_programs_metadata FOR EACH ROW EXECUTE PROCEDURE audit_function();
CREATE TRIGGER audit_trigger_groups BEFORE INSERT OR UPDATE OR DELETE ON groups FOR EACH ROW EXECUTE PROCEDURE audit_function();
CREATE TRIGGER audit_trigger_policies BEFORE INSERT OR UPDATE OR DELETE ON policies FOR EACH ROW EXECUTE PROCEDURE audit_function();
CREATE TRIGGER audit_trigger_programs BEFORE INSERT OR UPDATE OR DELETE ON programs FOR EACH ROW EXECUTE PROCEDURE audit_function();
CREATE TRIGGER audit_trigger_programs_groups_policies BEFORE INSERT OR UPDATE OR DELETE ON programs_groups_policies FOR EACH ROW EXECUTE PROCEDURE audit_function();
CREATE TRIGGER audit_trigger_recipients BEFORE INSERT OR UPDATE OR DELETE ON recipients FOR EACH ROW EXECUTE PROCEDURE audit_function();
CREATE TRIGGER audit_trigger_reports BEFORE INSERT OR UPDATE OR DELETE ON reports FOR EACH ROW EXECUTE PROCEDURE audit_function();
CREATE TRIGGER audit_trigger_teams BEFORE INSERT OR UPDATE OR DELETE ON teams FOR EACH ROW EXECUTE PROCEDURE audit_function();
CREATE TRIGGER audit_trigger_user_team BEFORE INSERT OR UPDATE OR DELETE ON user_team FOR EACH ROW EXECUTE PROCEDURE audit_function();
CREATE TRIGGER audit_trigger_users BEFORE INSERT OR UPDATE OR DELETE ON users FOR EACH ROW EXECUTE PROCEDURE audit_function();
