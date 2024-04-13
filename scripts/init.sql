DO $$
DECLARE
    i INT := 1;
BEGIN
    WHILE i <= 10 LOOP
        INSERT INTO feature DEFAULT VALUES;
        INSERT INTO tag DEFAULT VALUES;
        i := i + 1;
    END LOOP;
END $$;
