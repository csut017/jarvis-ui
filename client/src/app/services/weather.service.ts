import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { LoggingService, Logger } from './logging.service';
import { environment } from 'src/environments/environment';
import { Observable } from 'rxjs';
import { catchError, tap } from 'rxjs/operators';

@Injectable({
  providedIn: 'root'
})
export class WeatherService {

  constructor(private http: HttpClient,
    logging: LoggingService) {
    this.logger = logging.get('WeatherService');
  }

  private logger: Logger;

  getWeather(): Observable<WeatherSummary> {
    const url = environment.apiURL + `weather`;
    this.logger.log('Retrieving current weather information', url);
    return this.http.get<WeatherSummary>(url)
      .pipe(
        tap(res => this.logger.log('Retrieved weather', res)),
        catchError(this.logger.handleError<WeatherSummary>(`getWeather()`))
      );
  }
}

export interface WeatherSummary {
  current: string;
  oneWord: string;
  forecast: string;
}